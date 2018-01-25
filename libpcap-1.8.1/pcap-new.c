/*
 * Copyright (c) 2002 - 2005 NetGroup, Politecnico di Torino (Italy)
 * Copyright (c) 2005 - 2008 CACE Technologies, Davis (California)
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 * 1. Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 * notice, this list of conditions and the following disclaimer in the
 * documentation and/or other materials provided with the distribution.
 * 3. Neither the name of the Politecnico di Torino, CACE Technologies
 * nor the names of its contributors may be used to endorse or promote
 * products derived from this software without specific prior written
 * permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#include "pcap-int.h"	// for the details of the pcap_t structure
#include "pcap-rpcap.h"
#include "sockutils.h"
#include <errno.h>		// for the errno variable
#include <stdlib.h>		// for malloc(), free(), ...
#include <string.h>		// for strstr, etc

#ifndef WIN32
#include <dirent.h>		// for readdir
#endif

/* Keeps a list of all the opened connections in the active mode. */
extern struct activehosts *activeHosts;

/*
 * \brief Keeps the main socket identifier when we want to accept a new remote connection (active mode only).
 * See the documentation of pcap_remoteact_accept() and pcap_remoteact_cleanup() for more details.
 */
SOCKET sockmain;

/* String identifier to be used in the pcap_findalldevs_ex() */
#define PCAP_TEXT_SOURCE_FILE "File"
/* String identifier to be used in the pcap_findalldevs_ex() */
#define PCAP_TEXT_SOURCE_ADAPTER "Network adapter"

/* String identifier to be used in the pcap_findalldevs_ex() */
#define PCAP_TEXT_SOURCE_ON_LOCAL_HOST "on local host"
/* String identifier to be used in the pcap_findalldevs_ex() */
#define PCAP_TEXT_SOURCE_ON_REMOTE_HOST "on remote node"

/*
* Private data for capturing on WinPcap devices.
*/
struct pcap_win {
	int nonblock;
	int rfmon_selfstart;		/* a flag tells whether the monitor mode is set by itself */
	int filtering_in_kernel;	/* using kernel filter */

#ifdef HAVE_DAG_API
	int	dag_fcs_bits;		/* Number of checksum bits from link layer */
#endif
};

/****************************************************
 *                                                  *
 * Function bodies                                  *
 *                                                  *
 ****************************************************/

int pcap_findalldevs_ex(char *source, struct pcap_rmtauth *auth, pcap_if_t **alldevs, char *errbuf)
{
	SOCKET sockctrl;		/* socket descriptor of the control connection */
	uint32 totread = 0;		/* number of bytes of the payload read from the socket */
	int nread;
	struct addrinfo hints;		/* temp variable needed to resolve hostnames into to socket representation */
	struct addrinfo *addrinfo;	/* temp variable needed to resolve hostnames into to socket representation */
	struct rpcap_header header;	/* structure that keeps the general header of the rpcap protocol */
	int i, j;		/* temp variables */
	int naddr;		/* temp var needed to avoid problems with IPv6 addresses */
	struct pcap_addr *addr;	/* another such temp */
	int retval;		/* store the return value of the functions */
	int nif;		/* Number of interfaces listed */
	int active = 0;	/* 'true' if we the other end-party is in active mode */
	char host[PCAP_BUF_SIZE], port[PCAP_BUF_SIZE], name[PCAP_BUF_SIZE], path[PCAP_BUF_SIZE], filename[PCAP_BUF_SIZE];
	int type;
	pcap_t *fp;
	char tmpstring[PCAP_BUF_SIZE + 1];		/* Needed to convert names and descriptions from 'old' syntax to the 'new' one */
	pcap_if_t *dev;		/* Previous device into the pcap_if_t chain */


	if (strlen(source) > PCAP_BUF_SIZE)
	{
		pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The source string is too long. Cannot handle it correctly.");
		return -1;
	}

	/*
	 * Determine the type of the source (file, local, remote)
	 * There are some differences if pcap_findalldevs_ex() is called to list files and remote adapters.
	 * In the first case, the name of the directory we have to look into must be present (therefore
	 * the 'name' parameter of the pcap_parsesrcstr() is present).
	 * In the second case, the name of the adapter is not required (we need just the host). So, we have
	 * to use a first time this function to get the source type, and a second time to get the appropriate
	 * info, which depends on the source type.
	 */
	if (pcap_parsesrcstr(source, &type, NULL, NULL, NULL, errbuf) == -1)
		return -1;

	if (type == PCAP_SRC_IFLOCAL)
	{
		if (pcap_parsesrcstr(source, &type, host, NULL, NULL, errbuf) == -1)
			return -1;

		/* Initialize temporary string */
		tmpstring[PCAP_BUF_SIZE] = 0;

		/* The user wants to retrieve adapters from a local host */
		if (pcap_findalldevs(alldevs, errbuf) == -1)
			return -1;

		if ((alldevs == NULL) || (*alldevs == NULL))
		{
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
				"No interfaces found! Make sure libpcap/WinPcap is properly installed"
				" on the local machine.");
			return -1;
		}

		/* Scan all the interfaces and modify name and description */
		/* This is a trick in order to avoid the re-implementation of the pcap_findalldevs here */
		dev = *alldevs;
		while (dev)
		{
			/* Create the new device identifier */
			if (pcap_createsrcstr(tmpstring, PCAP_SRC_IFLOCAL, NULL, NULL, dev->name, errbuf) == -1)
				return -1;

			/* Delete the old pointer */
			free(dev->name);

			/* Make a copy of the new device identifier */
			dev->name = strdup(tmpstring);
			if (dev->name == NULL)
			{
				pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
				return -1;
			}

			/* Create the new device description */
			if ((dev->description == NULL) || (dev->description[0] == 0))
				pcap_snprintf(tmpstring, sizeof(tmpstring) - 1, "%s '%s' %s", PCAP_TEXT_SOURCE_ADAPTER,
				dev->name, PCAP_TEXT_SOURCE_ON_LOCAL_HOST);
			else
				pcap_snprintf(tmpstring, sizeof(tmpstring) - 1, "%s '%s' %s", PCAP_TEXT_SOURCE_ADAPTER,
				dev->description, PCAP_TEXT_SOURCE_ON_LOCAL_HOST);

			/* Delete the old pointer */
			free(dev->description);

			/* Make a copy of the description */
			dev->description = strdup(tmpstring);
			if (dev->description == NULL)
			{
				pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
				return -1;
			}

			dev = dev->next;
		}

		return 0;
	}

	(*alldevs) = NULL;

	if (type == PCAP_SRC_FILE)
	{
		size_t stringlen;
#ifdef WIN32
		WIN32_FIND_DATA filedata;
		HANDLE filehandle;
#else
		struct dirent *filedata;
		DIR *unixdir;
#endif

		if (pcap_parsesrcstr(source, &type, host, port, name, errbuf) == -1)
			return -1;

		/* Check that the filename is correct */
		stringlen = strlen(name);

		/* The directory must end with '\' in Win32 and '/' in UNIX */
#ifdef WIN32
#define ENDING_CHAR '\\'
#else
#define ENDING_CHAR '/'
#endif

		if (name[stringlen - 1] != ENDING_CHAR)
		{
			name[stringlen] = ENDING_CHAR;
			name[stringlen + 1] = 0;

			stringlen++;
		}

		/* Save the path for future reference */
		pcap_snprintf(path, sizeof(path), "%s", name);

#ifdef WIN32
		/* To perform directory listing, Win32 must have an 'asterisk' as ending char */
		if (name[stringlen - 1] != '*')
		{
			name[stringlen] = '*';
			name[stringlen + 1] = 0;
		}

		filehandle = FindFirstFile(name, &filedata);

		if (filehandle == INVALID_HANDLE_VALUE)
		{
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "Error when listing files: does folder '%s' exist?", path);
			return -1;
		}

#else
		/* opening the folder */
		unixdir= opendir(path);

		/* get the first file into it */
		filedata= readdir(unixdir);

		if (filedata == NULL)
		{
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "Error when listing files: does folder '%s' exist?", path);
			return -1;
		}
#endif

		do
		{

#ifdef WIN32
			pcap_snprintf(filename, sizeof(filename), "%s%s", path, filedata.cFileName);
#else
			pcap_snprintf(filename, sizeof(filename), "%s%s", path, filedata->d_name);
#endif

			fp = pcap_open_offline(filename, errbuf);

			if (fp)
			{
				/* allocate the main structure */
				if (*alldevs == NULL)	/* This is in case it is the first file */
				{
					(*alldevs) = (pcap_if_t *)malloc(sizeof(pcap_if_t));
					dev = (*alldevs);
				}
				else
				{
					dev->next = (pcap_if_t *)malloc(sizeof(pcap_if_t));
					dev = dev->next;
				}

				/* check that the malloc() didn't fail */
				if (dev == NULL)
				{
					pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
					return -1;
				}

				/* Initialize the structure to 'zero' */
				memset(dev, 0, sizeof(pcap_if_t));

				/* Create the new source identifier */
				if (pcap_createsrcstr(tmpstring, PCAP_SRC_FILE, NULL, NULL, filename, errbuf) == -1)
					return -1;

				stringlen = strlen(tmpstring);

				dev->name = (char *)malloc(stringlen + 1);
				if (dev->name == NULL)
				{
					pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
					return -1;
				}

				strlcpy(dev->name, tmpstring, stringlen);

				dev->name[stringlen] = 0;

				/* Create the description */
				pcap_snprintf(tmpstring, sizeof(tmpstring) - 1, "%s '%s' %s", PCAP_TEXT_SOURCE_FILE,
					filename, PCAP_TEXT_SOURCE_ON_LOCAL_HOST);

				stringlen = strlen(tmpstring);

				dev->description = (char *)malloc(stringlen + 1);

				if (dev->description == NULL)
				{
					pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
					return -1;
				}

				/* Copy the new device description into the correct memory location */
				strlcpy(dev->description, tmpstring, stringlen + 1);

				pcap_close(fp);
			}
		}
#ifdef WIN32
		while (FindNextFile(filehandle, &filedata) != 0);
#else
		while ( (filedata= readdir(unixdir)) != NULL);
#endif


#ifdef WIN32
		/* Close the search handle. */
		FindClose(filehandle);
#endif

		return 0;
	}

	/* If we come here, it is a remote host */

	/* Retrieve the needed data for getting adapter list */
	if (pcap_parsesrcstr(source, &type, host, port, NULL, errbuf) == -1)
		return -1;

	/* Warning: this call can be the first one called by the user. */
	/* For this reason, we have to initialize the WinSock support. */
	if (sock_init(errbuf, PCAP_ERRBUF_SIZE) == -1)
		return -1;

	/* Check for active mode */
	sockctrl = rpcap_remoteact_getsock(host, &active, errbuf);
	if (sockctrl == INVALID_SOCKET)
		return -1;

	if (!active) {
		/*
		 * We're not in active mode; let's try to open a new
		 * control connection.
		 */
		addrinfo = NULL;

		memset(&hints, 0, sizeof(struct addrinfo));
		hints.ai_family = PF_UNSPEC;
		hints.ai_socktype = SOCK_STREAM;

		if (port[0] == 0)
		{
			/* the user chose not to specify the port */
			if (sock_initaddress(host, RPCAP_DEFAULT_NETPORT, &hints, &addrinfo, errbuf, PCAP_ERRBUF_SIZE) == -1)
				return -1;
		}
		else
		{
			if (sock_initaddress(host, port, &hints, &addrinfo, errbuf, PCAP_ERRBUF_SIZE) == -1)
				return -1;
		}

		if ((sockctrl = sock_open(addrinfo, SOCKOPEN_CLIENT, 0, errbuf, PCAP_ERRBUF_SIZE)) == -1)
			goto error;

		/* addrinfo is no longer used */
		freeaddrinfo(addrinfo);
		addrinfo = NULL;

		if (rpcap_sendauth(sockctrl, auth, errbuf) == -1)
		{
			sock_close(sockctrl, NULL, 0);
			return -1;
		}
	}

	/* RPCAP findalldevs command */
	rpcap_createhdr(&header, RPCAP_MSG_FINDALLIF_REQ, 0, 0);

	if (sock_send(sockctrl, (char *)&header, sizeof(struct rpcap_header), errbuf, PCAP_ERRBUF_SIZE) == -1)
		goto error;

	if (sock_recv(sockctrl, (char *)&header, sizeof(struct rpcap_header), SOCK_RECEIVEALL_YES, errbuf, PCAP_ERRBUF_SIZE) == -1)
		goto error;

	/* Checks if the message is correct */
	retval = rpcap_checkmsg(errbuf, sockctrl, &header, RPCAP_MSG_FINDALLIF_REPLY, RPCAP_MSG_ERROR, 0);

	if (retval != RPCAP_MSG_FINDALLIF_REPLY)		/* the message is not the one expected */
	{
		switch (retval)
		{
		case -3:	/* Unrecoverable network error */
		case -2:	/* The other endpoint send a message that is not allowed here */
		case -1:	/* The other endpoint has a version number that is not compatible with our */
			break;

		case RPCAP_MSG_ERROR:		/* The other endpoint reported an error */
			break;

		default:
		{
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "Internal error");
			break;
		};
		}

		if (!active)
			sock_close(sockctrl, NULL, 0);

		return -1;
	}

	/* read the number of interfaces */
	nif = ntohs(header.value);

	/* loop until all interfaces have been received */
	for (i = 0; i < nif; i++)
	{
		struct rpcap_findalldevs_if findalldevs_if;
		char tmpstring2[PCAP_BUF_SIZE + 1];		/* Needed to convert names and descriptions from 'old' syntax to the 'new' one */
		size_t stringlen;

		tmpstring2[PCAP_BUF_SIZE] = 0;

		/* receive the findalldevs structure from remote host */
		nread = sock_recv(sockctrl, (char *)&findalldevs_if,
		    sizeof(struct rpcap_findalldevs_if), SOCK_RECEIVEALL_YES,
		    errbuf, PCAP_ERRBUF_SIZE);
		if (nread == -1)
			goto error;
		totread += nread;

		findalldevs_if.namelen = ntohs(findalldevs_if.namelen);
		findalldevs_if.desclen = ntohs(findalldevs_if.desclen);
		findalldevs_if.naddr = ntohs(findalldevs_if.naddr);

		/* allocate the main structure */
		if (i == 0)
		{
			(*alldevs) = (pcap_if_t *)malloc(sizeof(pcap_if_t));
			dev = (*alldevs);
		}
		else
		{
			dev->next = (pcap_if_t *)malloc(sizeof(pcap_if_t));
			dev = dev->next;
		}

		/* check that the malloc() didn't fail */
		if (dev == NULL)
		{
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
			goto error;
		}

		/* Initialize the structure to 'zero' */
		memset(dev, 0, sizeof(pcap_if_t));

		/* allocate mem for name and description */
		if (findalldevs_if.namelen)
		{

			if (findalldevs_if.namelen >= sizeof(tmpstring))
			{
				pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "Interface name too long");
				goto error;
			}

			/* Retrieve adapter name */
			nread = sock_recv(sockctrl, tmpstring,
			    findalldevs_if.namelen, SOCK_RECEIVEALL_YES,
			    errbuf, PCAP_ERRBUF_SIZE);
			if (nread == -1)
				goto error;
			totread += nread;

			tmpstring[findalldevs_if.namelen] = 0;

			/* Create the new device identifier */
			if (pcap_createsrcstr(tmpstring2, PCAP_SRC_IFREMOTE, host, port, tmpstring, errbuf) == -1)
				return -1;

			stringlen = strlen(tmpstring2);

			dev->name = (char *)malloc(stringlen + 1);
			if (dev->name == NULL)
			{
				pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
				goto error;
			}

			/* Copy the new device name into the correct memory location */
			strlcpy(dev->name, tmpstring2, stringlen + 1);
		}

		if (findalldevs_if.desclen)
		{
			if (findalldevs_if.desclen >= sizeof(tmpstring))
			{
				pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "Interface description too long");
				goto error;
			}

			/* Retrieve adapter description */
			nread = sock_recv(sockctrl, tmpstring,
			    findalldevs_if.desclen, SOCK_RECEIVEALL_YES,
			    errbuf, PCAP_ERRBUF_SIZE);
			if (nread == -1)
				goto error;
			totread += nread;

			tmpstring[findalldevs_if.desclen] = 0;

			pcap_snprintf(tmpstring2, sizeof(tmpstring2) - 1, "%s '%s' %s %s", PCAP_TEXT_SOURCE_ADAPTER,
				tmpstring, PCAP_TEXT_SOURCE_ON_REMOTE_HOST, host);

			stringlen = strlen(tmpstring2);

			dev->description = (char *)malloc(stringlen + 1);

			if (dev->description == NULL)
			{
				pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
				goto error;
			}

			/* Copy the new device description into the correct memory location */
			strlcpy(dev->description, tmpstring2, stringlen + 1);
		}

		dev->flags = ntohl(findalldevs_if.flags);

		naddr = 0;
		addr = NULL;
		/* loop until all addresses have been received */
		for (j = 0; j < findalldevs_if.naddr; j++)
		{
			struct rpcap_findalldevs_ifaddr ifaddr;

			/* Retrieve the interface addresses */
			nread = sock_recv(sockctrl, (char *)&ifaddr,
			    sizeof(struct rpcap_findalldevs_ifaddr),
			    SOCK_RECEIVEALL_YES, errbuf, PCAP_ERRBUF_SIZE);
			if (nread == -1)
				goto error;
			totread += nread;

			/*
			 * WARNING libpcap bug: the address listing is
			 * available only for AF_INET.
			 *
			 * XXX - IPv6?
			 */
			if (ntohs(ifaddr.addr.ss_family) == AF_INET)
			{
				if (addr == NULL)
				{
					dev->addresses = (struct pcap_addr *) malloc(sizeof(struct pcap_addr));
					addr = dev->addresses;
				}
				else
				{
					addr->next = (struct pcap_addr *) malloc(sizeof(struct pcap_addr));
					addr = addr->next;
				}
				naddr++;

				if (addr == NULL)
				{
					pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
					goto error;
				}
				addr->next = NULL;

				if (rpcap_deseraddr((struct sockaddr_storage *) &ifaddr.addr,
					(struct sockaddr_storage **) &addr->addr, errbuf) == -1)
					goto error;
				if (rpcap_deseraddr((struct sockaddr_storage *) &ifaddr.netmask,
					(struct sockaddr_storage **) &addr->netmask, errbuf) == -1)
					goto error;
				if (rpcap_deseraddr((struct sockaddr_storage *) &ifaddr.broadaddr,
					(struct sockaddr_storage **) &addr->broadaddr, errbuf) == -1)
					goto error;
				if (rpcap_deseraddr((struct sockaddr_storage *) &ifaddr.dstaddr,
					(struct sockaddr_storage **) &addr->dstaddr, errbuf) == -1)
					goto error;

				if ((addr->addr == NULL) && (addr->netmask == NULL) &&
					(addr->broadaddr == NULL) && (addr->dstaddr == NULL))
				{
					free(addr);
					addr = NULL;
					if (naddr == 1)
						naddr = 0;	/* the first item of the list had NULL addresses */
				}
			}
		}
	}

	/* Checks if all the data has been read; if not, discard the data in excess */
	if (totread != ntohl(header.plen))
	{
		if (sock_discard(sockctrl, ntohl(header.plen) - totread, errbuf, PCAP_ERRBUF_SIZE) == 1)
			return -1;
	}

	/* Control connection has to be closed only in case the remote machine is in passive mode */
	if (!active)
	{
		/* DO not send RPCAP_CLOSE, since we did not open a pcap_t; no need to free resources */
		if (sock_close(sockctrl, errbuf, PCAP_ERRBUF_SIZE))
			return -1;
	}

	/* To avoid inconsistencies in the number of sock_init() */
	sock_cleanup();

	return 0;

error:
	/*
	 * In case there has been an error, I don't want to overwrite it with a new one
	 * if the following call fails. I want to return always the original error.
	 *
	 * Take care: this connection can already be closed when we try to close it.
	 * This happens because a previous error in the rpcapd, which requested to
	 * closed the connection. In that case, we already recognized that into the
	 * rpspck_isheaderok() and we already acknowledged the closing.
	 * In that sense, this call is useless here (however it is needed in case
	 * the client generates the error).
	 *
	 * Checks if all the data has been read; if not, discard the data in excess
	 */
	if (totread != ntohl(header.plen))
	{
		if (sock_discard(sockctrl, ntohl(header.plen) - totread, NULL, 0) == 1)
			return -1;
	}

	/* Control connection has to be closed only in case the remote machine is in passive mode */
	if (!active)
		sock_close(sockctrl, NULL, 0);

	/* To avoid inconsistencies in the number of sock_init() */
	sock_cleanup();

	return -1;
}

int pcap_createsrcstr(char *source, int type, const char *host, const char *port, const char *name, char *errbuf)
{
	switch (type)
	{
	case PCAP_SRC_FILE:
	{
		strlcpy(source, PCAP_SRC_FILE_STRING, PCAP_BUF_SIZE);
		if ((name) && (*name))
		{
			strlcat(source, name, PCAP_BUF_SIZE);
			return 0;
		}
		else
		{
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The file name cannot be NULL.");
			return -1;
		}
	}

	case PCAP_SRC_IFREMOTE:
	{
		strlcpy(source, PCAP_SRC_IF_STRING, PCAP_BUF_SIZE);
		if ((host) && (*host))
		{
			if ((strcspn(host, "aAbBcCdDeEfFgGhHjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ")) == strlen(host))
			{
				/* the host name does not contains alphabetic chars. So, it is a numeric address */
				/* In this case we have to include it between square brackets */
				strlcat(source, "[", PCAP_BUF_SIZE);
				strlcat(source, host, PCAP_BUF_SIZE);
				strlcat(source, "]", PCAP_BUF_SIZE);
			}
			else
				strlcat(source, host, PCAP_BUF_SIZE);

			if ((port) && (*port))
			{
				strlcat(source, ":", PCAP_BUF_SIZE);
				strlcat(source, port, PCAP_BUF_SIZE);
			}

			strlcat(source, "/", PCAP_BUF_SIZE);
		}
		else
		{
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The host name cannot be NULL.");
			return -1;
		}

		if ((name) && (*name))
			strlcat(source, name, PCAP_BUF_SIZE);

		return 0;
	}

	case PCAP_SRC_IFLOCAL:
	{
		strlcpy(source, PCAP_SRC_IF_STRING, PCAP_BUF_SIZE);

		if ((name) && (*name))
			strlcat(source, name, PCAP_BUF_SIZE);

		return 0;
	}

	default:
	{
		pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The interface type is not valid.");
		return -1;
	}
	}
}

int pcap_parsesrcstr(const char *source, int *type, char *host, char *port, char *name, char *errbuf)
{
	char *ptr;
	int ntoken;
	char tmpname[PCAP_BUF_SIZE];
	char tmphost[PCAP_BUF_SIZE];
	char tmpport[PCAP_BUF_SIZE];
	int tmptype;

	/* Initialization stuff */
	tmpname[0] = 0;
	tmphost[0] = 0;
	tmpport[0] = 0;

	if (host)
		*host = 0;
	if (port)
		*port = 0;
	if (name)
		*name = 0;

	/* Look for a 'rpcap://' identifier */
	if ((ptr = strstr(source, PCAP_SRC_IF_STRING)) != NULL)
	{
		if (strlen(PCAP_SRC_IF_STRING) == strlen(source))
		{
			/* The source identifier contains only the 'rpcap://' string. */
			/* So, this is a local capture. */
			*type = PCAP_SRC_IFLOCAL;
			return 0;
		}

		ptr += strlen(PCAP_SRC_IF_STRING);

		if (strchr(ptr, '[')) /* This is probably a numeric address */
		{
			ntoken = sscanf(ptr, "[%[1234567890:.]]:%[^/]/%s", tmphost, tmpport, tmpname);

			if (ntoken == 1)	/* probably the port is missing */
				ntoken = sscanf(ptr, "[%[1234567890:.]]/%s", tmphost, tmpname);

			tmptype = PCAP_SRC_IFREMOTE;
		}
		else
		{
			ntoken = sscanf(ptr, "%[^/:]:%[^/]/%s", tmphost, tmpport, tmpname);

			if (ntoken == 1)
			{
				/*
				 * This can be due to two reasons:
				 * - we want a remote capture, but the network port is missing
				 * - we want to do a local capture
				 * To distinguish between the two, we look for the '/' char
				 */
				if (strchr(ptr, '/'))
				{
					/* We're on a remote capture */
					sscanf(ptr, "%[^/]/%s", tmphost, tmpname);
					tmptype = PCAP_SRC_IFREMOTE;
				}
				else
				{
					/* We're on a local capture */
					if (*ptr)
						strlcpy(tmpname, ptr, PCAP_BUF_SIZE);

					/* Clean the host name, since it is a remote capture */
					/* NOTE: the host name has been assigned in the previous "ntoken= sscanf(...)" line */
					tmphost[0] = 0;

					tmptype = PCAP_SRC_IFLOCAL;
				}
			}
			else
				tmptype = PCAP_SRC_IFREMOTE;
		}

		if (host)
			strlcpy(host, tmphost, PCAP_BUF_SIZE);
		if (port)
			strlcpy(port, tmpport, PCAP_BUF_SIZE);
		if (type)
			*type = tmptype;

		if (name)
		{
			/*
			 * If the user wants the host name, but it cannot be located into the source string, return error
			 * However, if the user is not interested in the interface name (e.g. if we're called by
			 * pcap_findalldevs_ex(), which does not have interface name, do not return error
			 */
			if (tmpname[0])
			{
				strlcpy(name, tmpname, PCAP_BUF_SIZE);
			}
			else
			{
				if (errbuf)
					pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The interface name has not been specified in the source string.");

				return -1;
			}
		}

		return 0;
	}

	/* Look for a 'file://' identifier */
	if ((ptr = strstr(source, PCAP_SRC_FILE_STRING)) != NULL)
	{
		ptr += strlen(PCAP_SRC_FILE_STRING);
		if (*ptr)
		{
			if (name)
				strlcpy(name, ptr, PCAP_BUF_SIZE);

			if (type)
				*type = PCAP_SRC_FILE;

			return 0;
		}
		else
		{
			if (errbuf)
				pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The file name has not been specified in the source string.");

			return -1;
		}

	}

	/* Backward compatibility; the user didn't use the 'rpcap://, file://'  specifiers */
	if ((source) && (*source))
	{
		if (name)
			strlcpy(name, source, PCAP_BUF_SIZE);

		if (type)
			*type = PCAP_SRC_IFLOCAL;

		return 0;
	}
	else
	{
		if (errbuf)
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The interface name has not been specified in the source string.");

		return -1;
	}
};

pcap_t *pcap_open(const char *source, int snaplen, int flags, int read_timeout, struct pcap_rmtauth *auth, char *errbuf)
{
	char host[PCAP_BUF_SIZE], port[PCAP_BUF_SIZE], name[PCAP_BUF_SIZE];
	int type;
	pcap_t *fp;
	int result;

	if (strlen(source) > PCAP_BUF_SIZE)
	{
		pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The source string is too long. Cannot handle it correctly.");
		return NULL;
	}

	/* determine the type of the source (file, local, remote) */
	if (pcap_parsesrcstr(source, &type, host, port, name, errbuf) == -1)
		return NULL;


	switch (type)
	{
	case PCAP_SRC_FILE:
		fp = pcap_open_offline(name, errbuf);
		break;

	case PCAP_SRC_IFREMOTE:
		fp = pcap_create(source, errbuf);
		if (fp == NULL)
		{
			return NULL;
		}

		/*
		 * Although we already have host, port and iface, we prefer TO PASS only 'pars' to the
		 * pcap_open_remote() so that it has to call the pcap_parsesrcstr() again.
		 * This is less optimized, but much clearer.
		 */

		result = pcap_opensource_remote(fp, auth);

		if (result != 0)
		{
			pcap_close(fp);
			return NULL;
		}

		struct pcap_md *md;				/* structure used when doing a remote live capture */
		md = (struct pcap_md *) ((u_char*)fp->priv + sizeof(struct pcap_win));

		fp->snapshot = snaplen;
		fp->opt.timeout = read_timeout;
		md->rmt_flags = flags;
		break;

	case PCAP_SRC_IFLOCAL:

		fp = pcap_open_live(name, snaplen, (flags & PCAP_OPENFLAG_PROMISCUOUS), read_timeout, errbuf);

#ifdef WIN32
		/*
		 * these flags are supported on Windows only
		 */
		if (fp != NULL && fp->adapter != NULL)
		{
			/* disable loopback capture if requested */
			if (flags & PCAP_OPENFLAG_NOCAPTURE_LOCAL)
			{
				if (!PacketSetLoopbackBehavior(fp->adapter, NPF_DISABLE_LOOPBACK))
				{
					pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "Unable to disable the capture of loopback packets.");
					pcap_close(fp);
					return NULL;
				}
			}

			/* set mintocopy to zero if requested */
			if (flags & PCAP_OPENFLAG_MAX_RESPONSIVENESS)
			{
				if (!PacketSetMinToCopy(fp->adapter, 0))
				{
					pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "Unable to set max responsiveness.");
					pcap_close(fp);
					return NULL;
				}
			}
		}
#endif /* WIN32 */

		break;

	default:
		strlcpy(errbuf, "Source type not supported", PCAP_ERRBUF_SIZE);
		return NULL;
	}
	return fp;
}

struct pcap_samp *pcap_setsampling(pcap_t *p)
{
	struct pcap_md *md;				/* structure used when doing a remote live capture */

	md = (struct pcap_md *) ((u_char*)p->priv + sizeof(struct pcap_win));
	return &(md->rmt_samp);
}

SOCKET pcap_remoteact_accept(const char *address, const char *port, const char *hostlist, char *connectinghost, struct pcap_rmtauth *auth, char *errbuf)
{
	/* socket-related variables */
	struct addrinfo hints;			/* temporary struct to keep settings needed to open the new socket */
	struct addrinfo *addrinfo;		/* keeps the addrinfo chain; required to open a new socket */
	struct sockaddr_storage from;	/* generic sockaddr_storage variable */
	socklen_t fromlen;				/* keeps the length of the sockaddr_storage variable */
	SOCKET sockctrl;				/* keeps the main socket identifier */
	struct activehosts *temp, *prev;	/* temp var needed to scan he host list chain */

	*connectinghost = 0;		/* just in case */

	/* Prepare to open a new server socket */
	memset(&hints, 0, sizeof(struct addrinfo));
	/* WARNING Currently it supports only ONE socket family among ipv4 and IPv6  */
	hints.ai_family = AF_INET;		/* PF_UNSPEC to have both IPv4 and IPv6 server */
	hints.ai_flags = AI_PASSIVE;	/* Ready to a bind() socket */
	hints.ai_socktype = SOCK_STREAM;

	/* Warning: this call can be the first one called by the user. */
	/* For this reason, we have to initialize the WinSock support. */
	if (sock_init(errbuf, PCAP_ERRBUF_SIZE) == -1)
		return -1;

	/* Do the work */
	if ((port == NULL) || (port[0] == 0))
	{
		if (sock_initaddress(address, RPCAP_DEFAULT_NETPORT_ACTIVE, &hints, &addrinfo, errbuf, PCAP_ERRBUF_SIZE) == -1)
		{
			SOCK_ASSERT(errbuf, 1);
			return -2;
		}
	}
	else
	{
		if (sock_initaddress(address, port, &hints, &addrinfo, errbuf, PCAP_ERRBUF_SIZE) == -1)
		{
			SOCK_ASSERT(errbuf, 1);
			return -2;
		}
	}


	if ((sockmain = sock_open(addrinfo, SOCKOPEN_SERVER, 1, errbuf, PCAP_ERRBUF_SIZE)) == -1)
	{
		SOCK_ASSERT(errbuf, 1);
		return -2;
	}

	/* Connection creation */
	fromlen = sizeof(struct sockaddr_storage);

	sockctrl = accept(sockmain, (struct sockaddr *) &from, &fromlen);

	/* We're not using sock_close, since we do not want to send a shutdown */
	/* (which is not allowed on a non-connected socket) */
	closesocket(sockmain);
	sockmain = 0;

	if (sockctrl == -1)
	{
		sock_geterror("accept(): ", errbuf, PCAP_ERRBUF_SIZE);
		return -2;
	}

	/* Get the numeric for of the name of the connecting host */
	if (getnameinfo((struct sockaddr *) &from, fromlen, connectinghost, RPCAP_HOSTLIST_SIZE, NULL, 0, NI_NUMERICHOST))
	{
		sock_geterror("getnameinfo(): ", errbuf, PCAP_ERRBUF_SIZE);
		rpcap_senderror(sockctrl, errbuf, PCAP_ERR_REMOTEACCEPT, NULL);
		sock_close(sockctrl, NULL, 0);
		return -1;
	}

	/* checks if the connecting host is among the ones allowed */
	if (sock_check_hostlist((char *)hostlist, RPCAP_HOSTLIST_SEP, &from, errbuf, PCAP_ERRBUF_SIZE) < 0)
	{
		rpcap_senderror(sockctrl, errbuf, PCAP_ERR_REMOTEACCEPT, NULL);
		sock_close(sockctrl, NULL, 0);
		return -1;
	}

	/* Send authentication to the remote machine */
	if (rpcap_sendauth(sockctrl, auth, errbuf) == -1)
	{
		rpcap_senderror(sockctrl, errbuf, PCAP_ERR_REMOTEACCEPT, NULL);
		sock_close(sockctrl, NULL, 0);
		return -3;
	}

	/* Checks that this host does not already have a cntrl connection in place */

	/* Initialize pointers */
	temp = activeHosts;
	prev = NULL;

	while (temp)
	{
		/* This host already has an active connection in place, so I don't have to update the host list */
		if (sock_cmpaddr(&temp->host, &from) == 0)
			return sockctrl;

		prev = temp;
		temp = temp->next;
	}

	/* The host does not exist in the list; so I have to update the list */
	if (prev)
	{
		prev->next = (struct activehosts *) malloc(sizeof(struct activehosts));
		temp = prev->next;
	}
	else
	{
		activeHosts = (struct activehosts *) malloc(sizeof(struct activehosts));
		temp = activeHosts;
	}

	if (temp == NULL)
	{
		pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "malloc() failed: %s", pcap_strerror(errno));
		rpcap_senderror(sockctrl, errbuf, PCAP_ERR_REMOTEACCEPT, NULL);
		sock_close(sockctrl, NULL, 0);
		return -1;
	}

	memcpy(&temp->host, &from, fromlen);
	temp->sockctrl = sockctrl;
	temp->next = NULL;

	return sockctrl;
}

int pcap_remoteact_close(const char *host, char *errbuf)
{
	struct activehosts *temp, *prev;	/* temp var needed to scan the host list chain */
	struct addrinfo hints, *addrinfo, *ai_next;	/* temp var needed to translate between hostname to its address */
	int retval;

	temp = activeHosts;
	prev = NULL;

	/* retrieve the network address corresponding to 'host' */
	addrinfo = NULL;
	memset(&hints, 0, sizeof(struct addrinfo));
	hints.ai_family = PF_UNSPEC;
	hints.ai_socktype = SOCK_STREAM;

	retval = getaddrinfo(host, "0", &hints, &addrinfo);
	if (retval != 0)
	{
		pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "getaddrinfo() %s", gai_strerror(retval));
		return -1;
	}

	while (temp)
	{
		ai_next = addrinfo;
		while (ai_next)
		{
			if (sock_cmpaddr(&temp->host, (struct sockaddr_storage *) ai_next->ai_addr) == 0)
			{
				struct rpcap_header header;

				/* Close this connection */
				rpcap_createhdr(&header, RPCAP_MSG_CLOSE, 0, 0);

				/* I don't check for errors, since I'm going to close everything */
				sock_send(temp->sockctrl, (char *)&header, sizeof(struct rpcap_header), errbuf, PCAP_ERRBUF_SIZE);

				if (sock_close(temp->sockctrl, errbuf, PCAP_ERRBUF_SIZE))
				{
					/* To avoid inconsistencies in the number of sock_init() */
					sock_cleanup();

					return -1;
				}

				if (prev)
					prev->next = temp->next;
				else
					activeHosts = temp->next;

				freeaddrinfo(addrinfo);

				free(temp);

				/* To avoid inconsistencies in the number of sock_init() */
				sock_cleanup();

				return 0;
			}

			ai_next = ai_next->ai_next;
		}
		prev = temp;
		temp = temp->next;
	}

	if (addrinfo)
		freeaddrinfo(addrinfo);

	/* To avoid inconsistencies in the number of sock_init() */
	sock_cleanup();

	pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The host you want to close the active connection is not known");
	return -1;
}

void pcap_remoteact_cleanup(void)
{
	/* Very dirty, but it works */
	if (sockmain)
	{
		closesocket(sockmain);

		/* To avoid inconsistencies in the number of sock_init() */
		sock_cleanup();
	}

}

int pcap_remoteact_list(char *hostlist, char sep, int size, char *errbuf)
{
	struct activehosts *temp;	/* temp var needed to scan the host list chain */
	size_t len;
	char hoststr[RPCAP_HOSTLIST_SIZE + 1];

	temp = activeHosts;

	len = 0;
	*hostlist = 0;

	while (temp)
	{
		/*int sock_getascii_addrport(const struct sockaddr_storage *sockaddr, char *address, int addrlen, char *port, int portlen, int flags, char *errbuf, int errbuflen) */

		/* Get the numeric form of the name of the connecting host */
		if (sock_getascii_addrport((struct sockaddr_storage *) &temp->host, hoststr,
			RPCAP_HOSTLIST_SIZE, NULL, 0, NI_NUMERICHOST, errbuf, PCAP_ERRBUF_SIZE) != -1)
			/*	if (getnameinfo( (struct sockaddr *) &temp->host, sizeof (struct sockaddr_storage), hoststr, */
			/*		RPCAP_HOSTLIST_SIZE, NULL, 0, NI_NUMERICHOST) ) */
		{
			/*	sock_geterror("getnameinfo(): ", errbuf, PCAP_ERRBUF_SIZE); */
			return -1;
		}

		len = len + strlen(hoststr) + 1 /* the separator */;

		if ((size < 0) || (len >= (size_t)size))
		{
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "The string you provided is not able to keep "
				"the hostnames for all the active connections");
			return -1;
		}

		strlcat(hostlist, hoststr, PCAP_ERRBUF_SIZE);
		hostlist[len - 1] = sep;
		hostlist[len] = 0;

		temp = temp->next;
	}

	return 0;
}
