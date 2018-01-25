#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#include <sys/param.h>

#include <stdlib.h>
#include <string.h>
#include <errno.h>

#include <ctype.h>
#include <netinet/in.h>
#include <sys/mman.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

#include <snf.h>
#if SNF_VERSION_API >= 0x0003
#define SNF_HAVE_INJECT_API
#endif

#include "pcap-int.h"
#include "pcap-snf.h"

/*
 * Private data for capturing on SNF devices.
 */
struct pcap_snf {
	snf_handle_t snf_handle; /* opaque device handle */
	snf_ring_t   snf_ring;   /* opaque device ring handle */
#ifdef SNF_HAVE_INJECT_API
        snf_inject_t snf_inj;    /* inject handle, if inject is used */
#endif
        int          snf_timeout;
        int          snf_boardnum;
};

static int
snf_set_datalink(pcap_t *p, int dlt)
{
	p->linktype = dlt;
	return (0);
}

static int
snf_pcap_stats(pcap_t *p, struct pcap_stat *ps)
{
	struct snf_ring_stats stats;
	struct pcap_snf *snfps = p->priv;
	int rc;

	if ((rc = snf_ring_getstats(snfps->snf_ring, &stats))) {
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "snf_get_stats: %s",
			 pcap_strerror(rc));
		return -1;
	}
	ps->ps_recv = stats.ring_pkt_recv + stats.ring_pkt_overflow;
	ps->ps_drop = stats.ring_pkt_overflow;
	ps->ps_ifdrop = stats.nic_pkt_overflow + stats.nic_pkt_bad;
	return 0;
}

static void
snf_platform_cleanup(pcap_t *p)
{
	struct pcap_snf *ps = p->priv;

#ifdef SNF_HAVE_INJECT_API
        if (ps->snf_inj)
                snf_inject_close(ps->snf_inj);
#endif
	snf_ring_close(ps->snf_ring);
	snf_close(ps->snf_handle);
	pcap_cleanup_live_common(p);
}

static int
snf_getnonblock(pcap_t *p, char *errbuf)
{
	struct pcap_snf *ps = p->priv;

	return (ps->snf_timeout == 0);
}

static int
snf_setnonblock(pcap_t *p, int nonblock, char *errbuf)
{
	struct pcap_snf *ps = p->priv;

	if (nonblock)
		ps->snf_timeout = 0;
	else {
		if (p->opt.timeout <= 0)
			ps->snf_timeout = -1; /* forever */
		else
			ps->snf_timeout = p->opt.timeout;
	}
	return (0);
}

#define _NSEC_PER_SEC 1000000000

static inline
struct timeval
snf_timestamp_to_timeval(const int64_t ts_nanosec, const int tstamp_precision)
{
	struct timeval tv;
	long tv_nsec;

	if (ts_nanosec == 0)
		return (struct timeval) { 0, 0 };

	tv.tv_sec = ts_nanosec / _NSEC_PER_SEC;
	tv_nsec = (ts_nanosec % _NSEC_PER_SEC);

	/* libpcap expects tv_usec to be nanos if using nanosecond precision. */
	if (tstamp_precision == PCAP_TSTAMP_PRECISION_NANO)
		tv.tv_usec = tv_nsec;
	else
		tv.tv_usec = tv_nsec / 1000;

	return tv;
}

static int
snf_read(pcap_t *p, int cnt, pcap_handler callback, u_char *user)
{
	struct pcap_snf *ps = p->priv;
	struct pcap_pkthdr hdr;
	int i, flags, err, caplen, n;
	struct snf_recv_req req;
	int nonblock, timeout;

	if (!p)
		return -1;

	n = 0;
	timeout = ps->snf_timeout;
	while (n < cnt || PACKET_COUNT_IS_UNLIMITED(cnt)) {
		/*
		 * Has "pcap_breakloop()" been called?
		 */
		if (p->break_loop) {
			if (n == 0) {
				p->break_loop = 0;
				return (-2);
			} else {
				return (n);
			}
		}

		err = snf_ring_recv(ps->snf_ring, timeout, &req);

		if (err) {
			if (err == EBUSY || err == EAGAIN) {
				return (n);
			}
			else if (err == EINTR) {
				timeout = 0;
				continue;
			}
			else {
				pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "snf_read: %s",
				 	 pcap_strerror(err));
				return -1;
			}
		}

		caplen = req.length;
		if (caplen > p->snapshot)
			caplen = p->snapshot;

		if ((p->fcode.bf_insns == NULL) ||
		     bpf_filter(p->fcode.bf_insns, req.pkt_addr, req.length, caplen)) {
			hdr.ts = snf_timestamp_to_timeval(req.timestamp, p->opt.tstamp_precision);
			hdr.caplen = caplen;
			hdr.len = req.length;
			callback(user, &hdr, req.pkt_addr);
		}
		n++;

		/* After one successful packet is received, we won't block
		* again for that timeout. */
		if (timeout != 0)
			timeout = 0;
	}
	return (n);
}

static int
snf_setfilter(pcap_t *p, struct bpf_program *fp)
{
	if (!p)
		return -1;
	if (!fp) {
		strncpy(p->errbuf, "setfilter: No filter specified",
			sizeof(p->errbuf));
		return -1;
	}

	/* Make our private copy of the filter */

	if (install_bpf_program(p, fp) < 0)
		return -1;

	return (0);
}

static int
snf_inject(pcap_t *p, const void *buf _U_, size_t size _U_)
{
#ifdef SNF_HAVE_INJECT_API
	struct pcap_snf *ps = p->priv;
        int rc;
        if (ps->snf_inj == NULL) {
                rc = snf_inject_open(ps->snf_boardnum, 0, &ps->snf_inj);
                if (rc) {
                        pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
                                "snf_inject_open: %s", pcap_strerror(rc));
                        return (-1);
                }
        }

        rc = snf_inject_send(ps->snf_inj, -1, 0, buf, size);
        if (!rc) {
                return (size);
        }
        else {
                pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "snf_inject_send: %s",
                         pcap_strerror(rc));
                return (-1);
        }
#else
	strlcpy(p->errbuf, "Sending packets isn't supported with this snf version",
	    PCAP_ERRBUF_SIZE);
	return (-1);
#endif
}

static int
snf_activate(pcap_t* p)
{
	struct pcap_snf *ps = p->priv;
	char *device = p->opt.device;
	const char *nr = NULL;
	int err;
	int flags = -1, ring_id = -1;

	if (device == NULL) {
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
			 "device is NULL: %s", pcap_strerror(errno));
		return -1;
	}

	/* In Libpcap, we set pshared by default if NUM_RINGS is set to > 1.
	 * Since libpcap isn't thread-safe */
	if ((nr = getenv("SNF_FLAGS")) && *nr)
		flags = strtol(nr, NULL, 0);
	else if ((nr = getenv("SNF_NUM_RINGS")) && *nr && atoi(nr) > 1)
		flags = SNF_F_PSHARED;
	else
		nr = NULL;

	err = snf_open(ps->snf_boardnum,
			0, /* let SNF API parse SNF_NUM_RINGS, if set */
			NULL, /* default RSS, or use SNF_RSS_FLAGS env */
			0, /* default to SNF_DATARING_SIZE from env */
			flags, /* may want pshared */
			&ps->snf_handle);
	if (err != 0) {
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
			 "snf_open failed: %s", pcap_strerror(err));
		return -1;
	}

	if ((nr = getenv("SNF_PCAP_RING_ID")) && *nr) {
		ring_id = (int) strtol(nr, NULL, 0);
	}
	err = snf_ring_open_id(ps->snf_handle, ring_id, &ps->snf_ring);
	if (err != 0) {
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
			 "snf_ring_open_id(ring=%d) failed: %s",
			 ring_id, pcap_strerror(err));
		return -1;
	}

	if (p->opt.timeout <= 0)
		ps->snf_timeout = -1;
	else
		ps->snf_timeout = p->opt.timeout;

	err = snf_start(ps->snf_handle);
	if (err != 0) {
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
			 "snf_start failed: %s", pcap_strerror(err));
		return -1;
	}

	/*
	 * "select()" and "poll()" don't work on snf descriptors.
	 */
	p->selectable_fd = -1;
	p->linktype = DLT_EN10MB;
	p->read_op = snf_read;
	p->inject_op = snf_inject;
	p->setfilter_op = snf_setfilter;
	p->setdirection_op = NULL; /* Not implemented.*/
	p->set_datalink_op = snf_set_datalink;
	p->getnonblock_op = snf_getnonblock;
	p->setnonblock_op = snf_setnonblock;
	p->stats_op = snf_pcap_stats;
	p->cleanup_op = snf_platform_cleanup;
#ifdef SNF_HAVE_INJECT_API
        ps->snf_inj = NULL;
#endif
	return 0;
}

#define MAX_DESC_LENGTH 128
int
snf_findalldevs(pcap_if_t **devlistp, char *errbuf)
{
	pcap_if_t *devlist = NULL,*curdev,*prevdev;
	pcap_addr_t *curaddr;
	struct snf_ifaddrs *ifaddrs, *ifa;
	char desc[MAX_DESC_LENGTH];
	int ret;

	if (snf_init(SNF_VERSION_API))
		return (-1);

	if (snf_getifaddrs(&ifaddrs) || ifaddrs == NULL)
	{
		(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			"snf_getifaddrs: %s", pcap_strerror(errno));
		return (-1);
	}
	ifa = ifaddrs;
	while (ifa)
	{
		/*
		 * Allocate a new entry
		 */
		curdev = (pcap_if_t *)malloc(sizeof(pcap_if_t));
		if (curdev == NULL) {
		(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			"snf_findalldevs malloc: %s", pcap_strerror(errno));
			return (-1);
		}
		if (devlist == NULL) /* save first entry */
			devlist = curdev;
		else
			prevdev->next = curdev;
		/*
		 * Fill in the entry.
		 */
		curdev->next = NULL;
		curdev->name = strdup(ifa->snf_ifa_name);
		if (curdev->name == NULL) {
			(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			    "snf_findalldevs strdup: %s", pcap_strerror(errno));
			free(curdev);
			return (-1);
		}
		(void)pcap_snprintf(desc,MAX_DESC_LENGTH,"Myricom snf%d",
				ifa->snf_ifa_portnum);
		curdev->description = strdup(desc);
		if (curdev->description == NULL) {
			(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			"snf_findalldevs strdup1: %s", pcap_strerror(errno));
			free(curdev->name);
			free(curdev);
			return (-1);
		}
		curdev->addresses = NULL;
		curdev->flags = 0;

		curaddr = (pcap_addr_t *)malloc(sizeof(pcap_addr_t));
		if (curaddr == NULL) {
			(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			     "snf_findalldevs malloc1: %s", pcap_strerror(errno));
			free(curdev->description);
			free(curdev->name);
			free(curdev);
			return (-1);
		}
		curdev->addresses = curaddr;
		curaddr->next = NULL;
		curaddr->addr = (struct sockaddr*)malloc(sizeof(struct sockaddr_storage));
		if (curaddr->addr == NULL) {
			(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			    "malloc2: %s", pcap_strerror(errno));
			free(curdev->description);
			free(curdev->name);
			free(curaddr);
			free(curdev);
			return (-1);
		}
		curaddr->addr->sa_family = AF_INET;
		curaddr->netmask = NULL;
		curaddr->broadaddr = NULL;
		curaddr->dstaddr = NULL;
		curaddr->next = NULL;

		prevdev = curdev;
		ifa = ifa->snf_ifa_next;
	}
	snf_freeifaddrs(ifaddrs);
	*devlistp = devlist;

	/*
	 * There are no platform-specific devices since each device
	 * exists as a regular Ethernet device.
	 */
	return 0;
}

pcap_t *
snf_create(const char *device, char *ebuf, int *is_ours)
{
	pcap_t *p;
	int boardnum = -1;
	struct snf_ifaddrs *ifaddrs, *ifa;
	size_t devlen;
	struct pcap_snf *ps;

	if (snf_init(SNF_VERSION_API)) {
		/* Can't initialize the API, so no SNF devices */
		*is_ours = 0;
		return NULL;
	}

	/*
	 * Match a given interface name to our list of interface names, from
	 * which we can obtain the intended board number
	 */
	if (snf_getifaddrs(&ifaddrs) || ifaddrs == NULL) {
		/* Can't get SNF addresses */
		*is_ours = 0;
		return NULL;
	}
	devlen = strlen(device) + 1;
	ifa = ifaddrs;
	while (ifa) {
		if (!strncmp(device, ifa->snf_ifa_name, devlen)) {
			boardnum = ifa->snf_ifa_boardnum;
			break;
		}
		ifa = ifa->snf_ifa_next;
	}
	snf_freeifaddrs(ifaddrs);

	if (ifa == NULL) {
		/*
		 * If we can't find the device by name, support the name "snfX"
		 * and "snf10gX" where X is the board number.
		 */
		if (sscanf(device, "snf10g%d", &boardnum) != 1 &&
		    sscanf(device, "snf%d", &boardnum) != 1) {
			/* Nope, not a supported name */
			*is_ours = 0;
			return NULL;
		    }
	}

	/* OK, it's probably ours. */
	*is_ours = 1;

	p = pcap_create_common(ebuf, sizeof (struct pcap_snf));
	if (p == NULL)
		return NULL;
	ps = p->priv;

	/*
	 * We support microsecond and nanosecond time stamps.
	 */
	p->tstamp_precision_count = 2;
	p->tstamp_precision_list = malloc(2 * sizeof(u_int));
	if (p->tstamp_precision_list == NULL) {
		pcap_snprintf(ebuf, PCAP_ERRBUF_SIZE, "malloc: %s",
		    pcap_strerror(errno));
		pcap_close(p);
		return NULL;
	}
	p->tstamp_precision_list[0] = PCAP_TSTAMP_PRECISION_MICRO;
	p->tstamp_precision_list[1] = PCAP_TSTAMP_PRECISION_NANO;

	p->activate_op = snf_activate;
	ps->snf_boardnum = boardnum;
	return p;
}

#ifdef SNF_ONLY
/*
 * This libpcap build supports only SNF cards, not regular network
 * interfaces..
 */

/*
 * There are no regular interfaces, just DAG interfaces.
 */
int
pcap_platform_finddevs(pcap_if_t **alldevsp, char *errbuf)
{
	*alldevsp = NULL;
	return (0);
}

/*
 * Attempts to open a regular interface fail.
 */
pcap_t *
pcap_create_interface(const char *device, char *errbuf)
{
	pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
	    "This version of libpcap only supports SNF cards");
	return NULL;
}
#endif
