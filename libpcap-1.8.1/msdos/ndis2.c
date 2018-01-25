/*
 * Copyright (c) 1993,1994
 *      Texas A&M University.  All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software
 *    must display the following acknowledgement:
 *      This product includes software developed by Texas A&M University
 *      and its contributors.
 * 4. Neither the name of the University nor the names of its contributors
 *    may be used to endorse or promote products derived from this software
 *    without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE UNIVERSITY AND CONTRIBUTORS ``AS IS'' AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED.  IN NO EVENT SHALL THE UNIVERSITY OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
 * OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 *
 * Developers:
 *             David K. Hess, Douglas Lee Schales, David R. Safford
 *
 * Heavily modified for Metaware HighC + GNU C 2.8+
 *             Gisle Vanem 1998
 */

#include <stdio.h>
#include <stdlib.h>
#include <dos.h>
#include <io.h>
#include <fcntl.h>
#include <malloc.h>
#include <string.h>

#include "pcap-dos.h"
#include "pcap-int.h"
#include "msdos/ndis2.h"

#if defined(USE_NDIS2)

/*
 *  Packet buffer handling
 */
extern int     FreePktBuf  (PktBuf *buf);
extern int     EnquePktBuf (PktBuf *buf);
extern PktBuf* AllocPktBuf (void);

/*
 *  Various defines
 */
#define MAX_NUM_DEBUG_STRINGS 90
#define DEBUG_STRING_LENGTH   80
#define STACK_POOL_SIZE       6
#define STACK_SIZE            256

#define MEDIA_FDDI            1
#define MEDIA_ETHERNET        2
#define MEDIA_TOKEN           3

static int     startDebug     = 0;
static int     stopDebug      = 0;

static DWORD   droppedPackets = 0L;
static WORD    frameSize      = 0;
static WORD    headerSize     = 0;
static int     mediaType      = 0;
static char   *lastErr        = NULL;

static BYTE    debugStrings [MAX_NUM_DEBUG_STRINGS][DEBUG_STRING_LENGTH];
static BYTE   *freeStacks   [STACK_POOL_SIZE];
static int     freeStackPtr = STACK_POOL_SIZE - 1;

static ProtMan protManEntry = NULL;
static WORD    protManDS    = 0;
static volatile int xmitPending;

static struct _PktBuf        *txBufPending;
static struct _CardHandle    *handle;
static struct _CommonChars    common;
static struct _ProtocolChars  protChars;
static struct _ProtDispatch   lowerTable;

static struct _FailingModules failingModules;
static struct _BindingsList   bindings;

static struct {
         WORD  err_num;
         char *err_text;
       } ndis_errlist[] = {

  { ERR_SUCCESS,
    "The function completed successfully.\n"  },

  { ERR_WAIT_FOR_RELEASE,
    "The ReceiveChain completed successfully but the protocol has\n"
    "retained control of the buffer.\n"  },

  { ERR_REQUEST_QUEUED,
    "The current request has been queued.\n"  },

  { ERR_FRAME_NOT_RECOGNIZED,
    "Frame not recognized.\n"  },

  { ERR_FRAME_REJECTED,
    "Frame was discarded.\n"  },

  { ERR_FORWARD_FRAME,
    "Protocol wishes to forward frame to another protocol.\n"  },

  { ERR_OUT_OF_RESOURCE,
    "Out of resource.\n"  },

  { ERR_INVALID_PARAMETER,
    "Invalid parameter.\n"  },

  { ERR_INVALID_FUNCTION,
    "Invalid function.\n"  },

  { ERR_NOT_SUPPORTED,
    "Not supported.\n"  },

  { ERR_HARDWARE_ERROR,
    "Hardware error.\n"  },

  { ERR_TRANSMIT_ERROR,
    "The packet was not transmitted due to an error.\n"  },

  { ERR_NO_SUCH_DESTINATION,
    "Token ring packet was not recognized when transmitted.\n"  },

  { ERR_BUFFER_TOO_SMALL,
    "Provided buffer was too small.\n"  },

  { ERR_ALREADY_STARTED,
    "Network drivers already started.\n"  },

  { ERR_INCOMPLETE_BINDING,
    "Protocol driver could not complete its bindings.\n"  },

  { ERR_DRIVER_NOT_INITIALIZED,
    "MAC did not initialize properly.\n"  },

  { ERR_HARDWARE_NOT_FOUND,
    "Hardware not found.\n"  },

  { ERR_HARDWARE_FAILURE,
    "Hardware failure.\n"  },

  { ERR_CONFIGURATION_FAILURE,
    "Configuration failure.\n"  },

  { ERR_INTERRUPT_CONFLICT,
    "Interrupt conflict.\n"  },

  { ERR_INCOMPATIBLE_MAC,
    "The MAC is not compatible with the protocol.\n"  },

  { ERR_INITIALIZATION_FAILED,
    "Initialization failed.\n"  },

  { ERR_NO_BINDING,
    "Binding did not occur.\n"  },

  { ERR_NETWORK_MAY_NOT_BE_CONNECTED,
    "The network may not be connected to the adapter.\n"  },

  { ERR_INCOMPATIBLE_OS_VERSION,
    "The version of the operating system is incompatible with the protocol.\n"  },

  { ERR_ALREADY_REGISTERED,
    "The protocol is already registered.\n"  },

  { ERR_PATH_NOT_FOUND,
    "PROTMAN.EXE could not be found.\n"  },

  { ERR_INSUFFICIENT_MEMORY,
    "Insufficient memory.\n"  },

  { ERR_INFO_NOT_FOUND,
    "Protocol Mananger info structure is lost or corrupted.\n"  },

  { ERR_GENERAL_FAILURE,
    "General failure.\n"  }
};

/*
 *  Some handy macros
 */
#define PERROR(str)    printf("%s (%d): %s\n", __FILE__,__LINE__,str)
#define DEBUG_RING()   (debugStrings[stopDebug+1 == MAX_NUM_DEBUG_STRINGS ? \
                        stopDebug = 0 : ++stopDebug])

/*
 * needs rewrite for DOSX
 */
#define MAC_DISPATCH(hnd)  ((struct _MacUpperDispatch*)(hnd)->common->upperDispatchTable)
#define MAC_STATUS(hnd)    ((struct _MacStatusTable*)  (hnd)->common->serviceStatus)
#define MAC_CHAR(hnd)      ((struct _MacChars*)        (hnd)->common->serviceChars)

#ifdef NDIS_DEBUG
  #define DEBUG0(str)      printf (str)
  #define DEBUG1(fmt,a)    printf (fmt,a)
  #define DEBUG2(fmt,a,b)  printf (fmt,a,b)
  #define TRACE0(str)      sprintf (DEBUG_RING(),str)
  #define TRACE1(fmt,a)    sprintf (DEBUG_RING(),fmt,a)
#else
  #define DEBUG0(str)      ((void)0)
  #define DEBUG1(fmt,a)    ((void)0)
  #define DEBUG2(fmt,a,b)  ((void)0)
  #define TRACE0(str)      ((void)0)
  #define TRACE1(fmt,a)    ((void)0)
#endif

/*
 * This routine is called from both threads
 */
void NdisFreeStack (BYTE *aStack)
{
  GUARD();

  if (freeStackPtr == STACK_POOL_SIZE - 1)
     PERROR ("tried to free too many stacks");

  freeStacks[++freeStackPtr] = aStack;

  if (freeStackPtr == 0)
     TRACE0 ("freeStackPtr went positive\n");

  UNGUARD();
}

/*
 * This routine is called from callbacks to allocate local data
 */
BYTE *NdisAllocStack (void)
{
  BYTE *stack;

  GUARD();

  if (freeStackPtr < 0)
  {
    /* Ran out of stack buffers. Return NULL which will start
     * dropping packets
     */
    TRACE0 ("freeStackPtr went negative\n");
    stack = 0;
  }
  else
    stack = freeStacks[freeStackPtr--];

  UNGUARD();
  return (stack);
}

CALLBACK (NdisSystemRequest (DWORD param1, DWORD param2, WORD param3,
                             WORD opcode, WORD targetDS))
{
  static int            bindEntry = 0;
  struct _CommonChars  *macCommon;
  volatile WORD result;

  switch (opcode)
  {
    case REQ_INITIATE_BIND:
         macCommon = (struct _CommonChars*) param2;
         if (macCommon == NULL)
	 {
           printf ("There is an NDIS misconfiguration.\n");
           result = ERR_GENERAL_FAILURE;
	   break;
	 }
         DEBUG2 ("module name %s\n"
                 "module type %s\n",
                 macCommon->moduleName,
                 ((MacChars*) macCommon->serviceChars)->macName);

         /* Binding to the MAC */
         result = macCommon->systemRequest ((DWORD)&common, (DWORD)&macCommon,
                                            0, REQ_BIND,
                                            macCommon->moduleDS);

         if (!strcmp(bindings.moduleName[bindEntry], handle->moduleName))
              handle->common = macCommon;
         else PERROR ("unknown module");
         ++bindEntry;
	 break;

    case REQ_INITIATE_UNBIND:
         macCommon = (struct _CommonChars*) param2;
         result = macCommon->systemRequest ((DWORD)&common, 0,
                                            0, REQ_UNBIND,
                                            macCommon->moduleDS);
         break;

    default:
         result = ERR_GENERAL_FAILURE;
	 break;
  }
  ARGSUSED (param1);
  ARGSUSED (param3);
  ARGSUSED (targetDS);
  return (result);
}

CALLBACK (NdisRequestConfirm (WORD protId, WORD macId,   WORD reqHandle,
                              WORD status, WORD request, WORD protDS))
{
  ARGSUSED (protId);    ARGSUSED (macId);
  ARGSUSED (reqHandle); ARGSUSED (status);
  ARGSUSED (request);   ARGSUSED (protDS);
  return (ERR_SUCCESS);
}

CALLBACK (NdisTransmitConfirm (WORD protId, WORD macId, WORD reqHandle,
                               WORD status, WORD protDS))
{
  xmitPending--;
  FreePktBuf (txBufPending);  /* Add passed ECB back to the free list */

  ARGSUSED (reqHandle);
  ARGSUSED (status);
  ARGSUSED (protDS);
  return (ERR_SUCCESS);
}


/*
 * The primary function for receiving packets
 */
CALLBACK (NdisReceiveLookahead (WORD  macId,      WORD  frameSize,
                                WORD  bytesAvail, BYTE *buffer,
                                BYTE *indicate,   WORD  protDS))
{
  int     result;
  PktBuf *pktBuf;
  WORD    bytesCopied;
  struct _TDBufDescr tDBufDescr;

#if 0
  TRACE1 ("lookahead length = %d, ", bytesAvail);
  TRACE1 ("ecb = %08lX, ",          *ecb);
  TRACE1 ("count = %08lX\n",         count);
  TRACE1 ("offset = %08lX, ",        offset);
  TRACE1 ("timesAllowed = %d, ",     timesAllowed);
  TRACE1 ("packet size = %d\n",      look->dataLookAheadLen);
#endif

  /* Allocate a buffer for the packet
   */
  if ((pktBuf = AllocPktBuf()) == NULL)
  {
    droppedPackets++;
    return (ERR_FRAME_REJECTED);
  }

  /*
   * Now kludge things. Note we will have to undo this later. This will
   * make the packet contiguous after the MLID has done the requested copy.
   */

  tDBufDescr.tDDataCount = 1;
  tDBufDescr.tDBufDescrRec[0].tDPtrType = NDIS_PTR_PHYSICAL;
  tDBufDescr.tDBufDescrRec[0].tDDataPtr = pktBuf->buffer;
  tDBufDescr.tDBufDescrRec[0].tDDataLen = pktBuf->length;
  tDBufDescr.tDBufDescrRec[0].dummy     = 0;

  result = MAC_DISPATCH(handle)->transferData (&bytesCopied, 0, &tDBufDescr,
                                               handle->common->moduleDS);
  pktBuf->packetLength = bytesCopied;

  if (result == ERR_SUCCESS)
       EnquePktBuf(pktBuf);
  else FreePktBuf (pktBuf);

  ARGSUSED (frameSize);
  ARGSUSED (bytesAvail);
  ARGSUSED (indicate);
  ARGSUSED (protDS);

  return (ERR_SUCCESS);
}

CALLBACK (NdisIndicationComplete (WORD macId, WORD protDS))
{
  ARGSUSED (macId);
  ARGSUSED (protDS);

  /* We don't give a hoot about these. Just return
   */
  return (ERR_SUCCESS);
}

/*
 * This is the OTHER way we may receive packets
 */
CALLBACK (NdisReceiveChain (WORD macId, WORD frameSize, WORD reqHandle,
                            struct _RxBufDescr *rxBufDescr,
                            BYTE *indicate, WORD protDS))
{
  struct _PktBuf *pktBuf;
  int     i;

  /*
   * For now we copy the entire packet over to a PktBuf structure. This may be
   * a performance hit but this routine probably isn't called very much, and
   * it is a lot of work to do it otherwise. Also if it is a filter protocol
   * packet we could end up sucking up MAC buffes.
   */

  if ((pktBuf = AllocPktBuf()) == NULL)
  {
    droppedPackets++;
    return (ERR_FRAME_REJECTED);
  }
  pktBuf->packetLength = 0;

  /* Copy the packet to the buffer
   */
  for (i = 0; i < rxBufDescr->rxDataCount; ++i)
  {
    struct _RxBufDescrRec *rxDescr = &rxBufDescr->rxBufDescrRec[i];

    memcpy (pktBuf->buffer + pktBuf->packetLength,
            rxDescr->rxDataPtr, rxDescr->rxDataLen);
    pktBuf->packetLength += rxDescr->rxDataLen;
  }

  EnquePktBuf (pktBuf);

  ARGSUSED (frameSize);
  ARGSUSED (reqHandle);
  ARGSUSED (indicate);
  ARGSUSED (protDS);

  /* This frees up the buffer for the MAC to use
   */
  return (ERR_SUCCESS);
}

CALLBACK (NdisStatusProc (WORD macId,  WORD param1, BYTE *indicate,
                          WORD opcode, WORD protDS))
{
  switch (opcode)
  {
    case STATUS_RING_STATUS:
	 break;
    case STATUS_ADAPTER_CHECK:
	 break;
    case STATUS_START_RESET:
	 break;
    case STATUS_INTERRUPT:
	 break;
    case STATUS_END_RESET:
	 break;
    default:
	 break;
  }
  ARGSUSED (macId);
  ARGSUSED (param1);
  ARGSUSED (indicate);
  ARGSUSED (opcode);
  ARGSUSED (protDS);

  /* We don't need to do anything about this stuff yet
   */
  return (ERR_SUCCESS);
}

/*
 * Tell the NDIS driver to start the delivery of the packet
 */
int NdisSendPacket (struct _PktBuf *pktBuf, int macId)
{
  struct _TxBufDescr txBufDescr;
  int     result;

  xmitPending++;
  txBufPending = pktBuf;    /* we only have 1 pending Tx at a time */

  txBufDescr.txImmedLen  = 0;
  txBufDescr.txImmedPtr  = NULL;
  txBufDescr.txDataCount = 1;
  txBufDescr.txBufDescrRec[0].txPtrType = NDIS_PTR_PHYSICAL;
  txBufDescr.txBufDescrRec[0].dummy     = 0;
  txBufDescr.txBufDescrRec[0].txDataLen = pktBuf->packetLength;
  txBufDescr.txBufDescrRec[0].txDataPtr = pktBuf->buffer;

  result = MAC_DISPATCH(handle)->transmitChain (common.moduleId,
                                                pktBuf->handle,
                                                &txBufDescr,
                                                handle->common->moduleDS);
  switch (result)
  {
    case ERR_OUT_OF_RESOURCE:
         /* Note that this should not happen but if it does there is not
          * much we can do about it
          */
         printf ("ERROR: transmit queue overflowed\n");
         return (0);

    case ERR_SUCCESS:
         /* Everything was hunky dory and synchronous. Free up the
          * packet buffer
          */
         xmitPending--;
         FreePktBuf (pktBuf);
         return (1);

    case ERR_REQUEST_QUEUED:
         /* Everything was hunky dory and asynchronous. Do nothing
          */
         return (1);

    default:
         printf ("Tx fail, code = %04X\n", result);
         return (0);
  }
}



static int ndis_nerr = sizeof(ndis_errlist) / sizeof(ndis_errlist[0]);

static char *Ndis_strerror (WORD errorCode)
{
  static char buf[30];
  int    i;

  for (i = 0; i < ndis_nerr; i++)
      if (errorCode == ndis_errlist[i].err_num)
         return (ndis_errlist[i].err_text);

  sprintf (buf,"unknown error %d",errorCode);
  return (buf);
}


char *NdisLastError (void)
{
  char *errStr = lastErr;
  lastErr = NULL;
  return (errStr);
}

int NdisOpen (void)
{
  struct _ReqBlock reqBlock;
  int     result;
  int     ndisFd = open (NDIS_PATH, O_RDONLY);

  if (ndisFd < 0)
  {
    printf ("Could not open NDIS Protocol Manager device.\n");
    return (0);
  }

  memset (&reqBlock, 0, sizeof(ReqBlock));

  reqBlock.opcode = PM_GET_PROTOCOL_MANAGER_LINKAGE;

  result = NdisGetLinkage (ndisFd, (char*)&reqBlock, sizeof(ReqBlock));
  if (result != 0)
  {
    printf ("Could not get Protocol Manager linkage.\n");
    close (ndisFd);
    return (0);
  }

  close (ndisFd);
  protManEntry = (ProtMan) reqBlock.pointer1;
  protManDS    = reqBlock.word1;

  DEBUG2 ("Entry Point = %04X:%04X\n", FP_SEG(protManEntry),FP_OFF(protManEntry));
  DEBUG1 ("ProtMan DS  = %04X\n", protManDS);
  return (1);
}


int NdisRegisterAndBind (int promis)
{
  struct _ReqBlock reqBlock;
  WORD    result;

  memset (&common,0,sizeof(common));

  common.tableSize = sizeof (common);

  common.majorNdisVersion   = 2;
  common.minorNdisVersion   = 0;
  common.majorModuleVersion = 2;
  common.minorModuleVersion = 0;

  /* Indicates binding from below and dynamically loaded
   */
  common.moduleFlags = 0x00000006L;

  strcpy (common.moduleName, "PCAP");

  common.protocolLevelUpper = 0xFF;
  common.protocolLevelLower = 1;
  common.interfaceLower     = 1;
#ifdef __DJGPP__
  common.moduleDS           = _dos_ds; /* the callback data segment */
#else
  common.moduleDS           = _DS;
#endif

  common.systemRequest      = (SystemRequest) systemRequestGlue;
  common.serviceChars       = (BYTE*) &protChars;
  common.serviceStatus      = NULL;
  common.upperDispatchTable = NULL;
  common.lowerDispatchTable = (BYTE*) &lowerTable;

  protChars.length  = sizeof (protChars);
  protChars.name[0] = 0;
  protChars.type    = 0;

  lowerTable.backPointer        = &common;
  lowerTable.requestConfirm     = requestConfirmGlue;
  lowerTable.transmitConfirm    = transmitConfirmGlue;
  lowerTable.receiveLookahead   = receiveLookaheadGlue;
  lowerTable.indicationComplete = indicationCompleteGlue;
  lowerTable.receiveChain       = receiveChainGlue;
  lowerTable.status             = statusGlue;
  lowerTable.flags              = 3;
  if (promis)
     lowerTable.flags |= 4;   /* promiscous mode (receive everything) */

  bindings.numBindings = 1;
  strcpy (bindings.moduleName[0], handle->moduleName);

  /* Register ourselves with NDIS
   */
  reqBlock.opcode   = PM_REGISTER_MODULE;
  reqBlock.pointer1 = (BYTE FAR*) &common;
  reqBlock.pointer2 = (BYTE FAR*) &bindings;

  result = (*protManEntry) (&reqBlock, protManDS);
  if (result)
  {
    printf ("Protman registering failed: %s\n", Ndis_strerror(result));
    return (0);
  }

  /* Start the binding process
   */
  reqBlock.opcode   = PM_BIND_AND_START;
  reqBlock.pointer1 = (BYTE FAR*) &failingModules;

  result = (*protManEntry) (&reqBlock, protManDS);
  if (result)
  {
    printf ("Start binding failed: %s\n", Ndis_strerror(result));
    return (0);
  }
  return (1);
}

static int CheckMacFeatures (CardHandle *card)
{
  DWORD      serviceFlags;
  BYTE _far *mediaString;
  BYTE _far *mac_addr;

  DEBUG2 ("checking card features\n"
          "common table address = %08lX, macId = %d\n",
          card->common, card->common->moduleId);

  serviceFlags = MAC_CHAR (handle)->serviceFlags;

  if ((serviceFlags & SF_PROMISCUOUS) == 0)
  {
    printf ("The MAC %s does not support promiscuous mode.\n",
            card->moduleName);
    return (0);
  }

  mediaString = MAC_CHAR (handle)->macName;

  DEBUG1 ("media type = %s\n",mediaString);

  /* Get the media type. And set the header size
   */
  if (!strncmp(mediaString,"802.3",5) ||
      !strncmp(mediaString,"DIX",3)   ||
      !strncmp(mediaString,"DIX+802.3",9))
       headerSize = sizeof (EthernetIIHeader);

  else if (!strncmp(mediaString,"FDDI",4))
       headerSize = sizeof (FddiHeader) +
                    sizeof (Ieee802Dot2SnapHeader);
  else
  {
    printf ("Unsupported MAC type: `%s'\n", mediaString);
    return (0);
  }

  frameSize = MAC_CHAR (handle)->maxFrameSize;
  mac_addr  = MAC_CHAR (handle)->currentAddress;

  printf ("Hardware address: %02X:%02X:%02X:%02X:%02X:%02X\n",
          mac_addr[0], mac_addr[1], mac_addr[2],
          mac_addr[3], mac_addr[4], mac_addr[5]);
  return (1);
}

static int NdisStartMac (CardHandle *card)
{
  WORD result;

  /* Set the lookahead length
   */
  result = MAC_DISPATCH(handle)->request (common.moduleId, 0,
                                          headerSize, 0,
                                          REQ_SET_LOOKAHEAD,
                                          card->common->moduleDS);

  /* We assume that if we got INVALID PARAMETER then either this
   * is not supported or will work anyway. NE2000 does this.
   */
  if (result != ERR_SUCCESS && result != ERR_INVALID_PARAMETER)
  {
    DEBUG1 ("Set lookahead failed: %s\n", Ndis_strerror(result));
    return (0);
  }

  /* Set the packet filter. Note that for some medias and drivers we
   * must specify all three flags or the card(s) will not operate correctly.
   */
  result = MAC_DISPATCH(handle)->request (common.moduleId, 0,
                      /* all packets */   FILTER_PROMISCUOUS |
                      /* packets to us */ FILTER_DIRECTED    |
                      /* broadcasts */    FILTER_BROADCAST,
                                          0, REQ_SET_PACKET_FILTER,
                                          card->common->moduleDS);
  if (result != ERR_SUCCESS)
  {
    DEBUG1 ("Set packet filter failed: %s\n", Ndis_strerror(result));
    return (0);
  }

  /* If OPEN/CLOSE supported then open the adapter
   */
  if (MAC_CHAR(handle)->serviceFlags & SF_OPEN_CLOSE)
  {
    result = MAC_DISPATCH(handle)->request (common.moduleId, 0, 0, NULL,
                                            REQ_OPEN_ADAPTER,
                                            card->common->moduleDS);
    if (result != ERR_SUCCESS)
    {
      DEBUG1 ("Opening the MAC failed: %s\n", Ndis_strerror(result));
      return (0);
    }
  }
  return (1);
}

void NdisShutdown (void)
{
  struct _ReqBlock reqBlock;
  int     result, i;

  if (!handle)
     return;

  /* If the adapters support open and are open then close them
   */
  if ((MAC_CHAR(handle)->serviceFlags & SF_OPEN_CLOSE) &&
      (MAC_STATUS(handle)->macStatus & MAC_OPEN))
  {
    result = MAC_DISPATCH(handle)->request (common.moduleId, 0, 0, 0,
                                            REQ_CLOSE_ADAPTER,
                                            handle->common->moduleDS);
    if (result != ERR_SUCCESS)
    {
      printf ("Closing the MAC failed: %s\n", Ndis_strerror(result));
      return;
    }
  }

  /* Tell the Protocol Manager to unbind and stop
   */
  reqBlock.opcode   = PM_UNBIND_AND_STOP;
  reqBlock.pointer1 = (BYTE FAR*) &failingModules;
  reqBlock.pointer2 = NULL;

  result = (*protManEntry) (&reqBlock, protManDS);
  if (result)
     printf ("Unbind failed: %s\n",  Ndis_strerror(result));

  for (i = 0; i < STACK_POOL_SIZE; ++i)
     free (freeStacks[i] - STACK_SIZE);

  handle = NULL;
}

int NdisInit (int promis)
{
  int i, result;

  /* Allocate the real mode stacks used for NDIS callbacks
   */
  for (i = 0; i < STACK_POOL_SIZE; ++i)
  {
    freeStacks[i] = malloc (STACK_SIZE);
    if (!freeStacks[i])
       return (0);
    freeStacks[i] += STACK_SIZE;
  }

  if (!NdisOpen())
     return (0);

  if (!NdisRegisterAndBind(promis))
     return (0);

  DEBUG1 ("My module id: %d\n", common.moduleId);
  DEBUG1 ("Handle id;    %d\n", handle->common->moduleId);
  DEBUG1 ("MAC card:     %-16s - ", handle->moduleName);

  atexit (NdisShutdown);

  if (!CheckMacFeatures(&handle))
     return (0);

  switch (mediaType)
  {
    case MEDIA_FDDI:
         DEBUG0 ("Media type: FDDI");
	 break;
    case MEDIA_ETHERNET:
         DEBUG0 ("Media type: ETHERNET");
	 break;
    default:
         DEBUG0 ("Unsupported media.\n");
         return (0);
  }

  DEBUG1 (" - Frame size: %d\n", frameSize);

  if (!NdisStartMac(&handle))
     return (0);
  return (1);
}
#endif  /* USE_NDIS2 */

