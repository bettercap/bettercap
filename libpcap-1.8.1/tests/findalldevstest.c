#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#include <stdlib.h>
#include <sys/types.h>
#ifdef _WIN32
  #include <winsock2.h>
#else
  #include <sys/socket.h>
  #include <netinet/in.h>
  #include <arpa/inet.h>
  #include <netdb.h>
#endif

#include <pcap.h>

static int ifprint(pcap_if_t *d);
static char *iptos(bpf_u_int32 in);

int main(int argc, char **argv)
{
  pcap_if_t *alldevs;
  pcap_if_t *d;
  char *s;
  bpf_u_int32 net, mask;
  int exit_status = 0;

  char errbuf[PCAP_ERRBUF_SIZE+1];
  if (pcap_findalldevs(&alldevs, errbuf) == -1)
  {
    fprintf(stderr,"Error in pcap_findalldevs: %s\n",errbuf);
    exit(1);
  }
  for(d=alldevs;d;d=d->next)
  {
    if (!ifprint(d))
      exit_status = 2;
  }

  if ( (s = pcap_lookupdev(errbuf)) == NULL)
  {
    fprintf(stderr,"Error in pcap_lookupdev: %s\n",errbuf);
    exit_status = 2;
  }
  else
  {
    printf("Preferred device name: %s\n",s);
  }

  if (pcap_lookupnet(s, &net, &mask, errbuf) < 0)
  {
    fprintf(stderr,"Error in pcap_lookupnet: %s\n",errbuf);
    exit_status = 2;
  }
  else
  {
    printf("Preferred device is on network: %s/%s\n",iptos(net), iptos(mask));
  }

  exit(exit_status);
}

static int ifprint(pcap_if_t *d)
{
  pcap_addr_t *a;
#ifdef INET6
  char ntop_buf[INET6_ADDRSTRLEN];
#endif
  const char *sep;
  int status = 1; /* success */

  printf("%s\n",d->name);
  if (d->description)
    printf("\tDescription: %s\n",d->description);
  printf("\tFlags: ");
  sep = "";
  if (d->flags & PCAP_IF_UP) {
    printf("%sUP", sep);
    sep = ", ";
  }
  if (d->flags & PCAP_IF_RUNNING) {
    printf("%sRUNNING", sep);
    sep = ", ";
  }
  if (d->flags & PCAP_IF_LOOPBACK) {
    printf("%sLOOPBACK", sep);
    sep = ", ";
  }
  printf("\n");

  for(a=d->addresses;a;a=a->next) {
    if (a->addr != NULL)
      switch(a->addr->sa_family) {
      case AF_INET:
        printf("\tAddress Family: AF_INET\n");
        if (a->addr)
          printf("\t\tAddress: %s\n",
            inet_ntoa(((struct sockaddr_in *)(a->addr))->sin_addr));
        if (a->netmask)
          printf("\t\tNetmask: %s\n",
            inet_ntoa(((struct sockaddr_in *)(a->netmask))->sin_addr));
        if (a->broadaddr)
          printf("\t\tBroadcast Address: %s\n",
            inet_ntoa(((struct sockaddr_in *)(a->broadaddr))->sin_addr));
        if (a->dstaddr)
          printf("\t\tDestination Address: %s\n",
            inet_ntoa(((struct sockaddr_in *)(a->dstaddr))->sin_addr));
        break;
#ifdef INET6
      case AF_INET6:
        printf("\tAddress Family: AF_INET6\n");
        if (a->addr)
          printf("\t\tAddress: %s\n",
            inet_ntop(AF_INET6,
               ((struct sockaddr_in6 *)(a->addr))->sin6_addr.s6_addr,
               ntop_buf, sizeof ntop_buf));
        if (a->netmask)
          printf("\t\tNetmask: %s\n",
            inet_ntop(AF_INET6,
              ((struct sockaddr_in6 *)(a->netmask))->sin6_addr.s6_addr,
               ntop_buf, sizeof ntop_buf));
        if (a->broadaddr)
          printf("\t\tBroadcast Address: %s\n",
            inet_ntop(AF_INET6,
              ((struct sockaddr_in6 *)(a->broadaddr))->sin6_addr.s6_addr,
               ntop_buf, sizeof ntop_buf));
        if (a->dstaddr)
          printf("\t\tDestination Address: %s\n",
            inet_ntop(AF_INET6,
              ((struct sockaddr_in6 *)(a->dstaddr))->sin6_addr.s6_addr,
               ntop_buf, sizeof ntop_buf));
        break;
#endif
      default:
        printf("\tAddress Family: Unknown (%d)\n", a->addr->sa_family);
        break;
      }
    else
    {
      fprintf(stderr, "\tWarning: a->addr is NULL, skipping this address.\n");
      status = 0;
    }
  }
  printf("\n");
  return status;
}

/* From tcptraceroute */
#define IPTOSBUFFERS	12
static char *iptos(bpf_u_int32 in)
{
	static char output[IPTOSBUFFERS][3*4+3+1];
	static short which;
	u_char *p;

	p = (u_char *)&in;
	which = (which + 1 == IPTOSBUFFERS ? 0 : which + 1);
	sprintf(output[which], "%d.%d.%d.%d", p[0], p[1], p[2], p[3]);
	return output[which];
}
