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

#ifndef __PCAP_NDIS_H
#define __PCAP_NDIS_H

#if defined (__HIGHC__)
  #define pascal          _CC(_CALLEE_POPS_STACK & ~_REVERSE_PARMS) /* calling convention */
  #define CALLBACK(foo)   pascal WORD foo
  #define PAS_PTR(x,arg)  typedef FAR WORD pascal (*x) arg
  #define GUARD()         _inline (0x9C,0xFA)   /* pushfd, cli */
  #define UNGUARD()       _inline (0x9D)        /* popfd */
  #define FAR             _far

#elif defined(__GNUC__)
  #define CALLBACK(foo)   WORD foo __attribute__((stdcall))
  #define PAS_PTR(x,arg)  typedef WORD (*x) arg __attribute__((stdcall))
  #define GUARD()         __asm__ __volatile__ ("pushfd; cli")
  #define UNGUARD()       __asm__ __volatile__ ("popfd")
  #define FAR

#elif defined (__TURBOC__)
  #define CALLBACK(foo)   WORD pascal foo
  #define PAS_PTR(x,arg)  typedef WORD pascal (_far *x) arg
  #define GUARD()         _asm { pushf; cli }
  #define UNGUARD()       _asm { popf }
  #define FAR             _far

#elif defined (__WATCOMC__)
  #define CALLBACK(foo)   WORD pascal foo
  #define PAS_PTR(x,arg)  typedef WORD pascal (_far *x) arg
  #define GUARD()         _disable()
  #define UNGUARD()       _enable()
  #define FAR             _far

#else
  #error Unsupported compiler
#endif


/*
 *  Forwards
 */
struct _ReqBlock;
struct _TxBufDescr;
struct _TDBufDescr;

/*
 * Protocol Manager API
 */
PAS_PTR (ProtMan, (struct _ReqBlock FAR*, WORD));

/*
 * System request
 */
PAS_PTR (SystemRequest, (DWORD, DWORD, WORD, WORD, WORD));

/*
 * MAC API
 */
PAS_PTR (TransmitChain, (WORD, WORD, struct _TxBufDescr FAR*, WORD));
PAS_PTR (TransferData,  (WORD*,WORD, struct _TDBufDescr FAR*, WORD));
PAS_PTR (Request,       (WORD, WORD, WORD, DWORD, WORD, WORD));
PAS_PTR (ReceiveRelease,(WORD, WORD));
PAS_PTR (IndicationOn,  (WORD));
PAS_PTR (IndicationOff, (WORD));


typedef enum {
        HARDWARE_NOT_INSTALLED  = 0,
        HARDWARE_FAILED_DIAG    = 1,
        HARDWARE_FAILED_CONFIG  = 2,
        HARDWARE_HARD_FAULT     = 3,
        HARDWARE_SOFT_FAULT     = 4,
        HARDWARE_OK             = 7,
        HARDWARE_MASK           = 0x0007,
        MAC_BOUND               = 0x0008,
        MAC_OPEN                = 0x0010,
        DIAG_IN_PROGRESS        = 0x0020
      } NdisMacStatus;

typedef enum {
        STATUS_RING_STATUS      = 1,
        STATUS_ADAPTER_CHECK    = 2,
        STATUS_START_RESET      = 3,
        STATUS_INTERRUPT        = 4,
        STATUS_END_RESET        = 5
      } NdisStatus;

typedef enum {
        FILTER_DIRECTED         = 1,
        FILTER_BROADCAST        = 2,
        FILTER_PROMISCUOUS      = 4,
        FILTER_SOURCE_ROUTE     = 8
      } NdisPacketFilter;

typedef enum {
        REQ_INITIATE_DIAGNOSTICS     = 1,
        REQ_READ_ERROR_LOG           = 2,
        REQ_SET_STATION_ADDRESS      = 3,
        REQ_OPEN_ADAPTER             = 4,
        REQ_CLOSE_ADAPTER            = 5,
        REQ_RESET_MAC                = 6,
        REQ_SET_PACKET_FILTER        = 7,
        REQ_ADD_MULTICAST_ADDRESS    = 8,
        REQ_DELETE_MULTICAST_ADDRESS = 9,
        REQ_UPDATE_STATISTICS        = 10,
        REQ_CLEAR_STATISTICS         = 11,
        REQ_INTERRUPT_REQUEST        = 12,
        REQ_SET_FUNCTIONAL_ADDRESS   = 13,
        REQ_SET_LOOKAHEAD            = 14
      } NdisGeneralRequest;

typedef enum {
        SF_BROADCAST             = 0x00000001L,
        SF_MULTICAST             = 0x00000002L,
        SF_FUNCTIONAL            = 0x00000004L,
        SF_PROMISCUOUS           = 0x00000008L,
        SF_SOFT_ADDRESS          = 0x00000010L,
        SF_STATS_CURRENT         = 0x00000020L,
        SF_INITIATE_DIAGS        = 0x00000040L,
        SF_LOOPBACK              = 0x00000080L,
        SF_RECEIVE_CHAIN         = 0x00000100L,
        SF_SOURCE_ROUTING        = 0x00000200L,
        SF_RESET_MAC             = 0x00000400L,
        SF_OPEN_CLOSE            = 0x00000800L,
        SF_INTERRUPT_REQUEST     = 0x00001000L,
        SF_SOURCE_ROUTING_BRIDGE = 0x00002000L,
        SF_VIRTUAL_ADDRESSES     = 0x00004000L
      } NdisMacServiceFlags;

typedef enum {
        REQ_INITIATE_BIND        = 1,
        REQ_BIND                 = 2,
        REQ_INITIATE_PREBIND     = 3,
        REQ_INITIATE_UNBIND      = 4,
        REQ_UNBIND               = 5
      } NdisSysRequest;

typedef enum  {
        PM_GET_PROTOCOL_MANAGER_INFO      = 1,
        PM_REGISTER_MODULE                = 2,
        PM_BIND_AND_START                 = 3,
        PM_GET_PROTOCOL_MANAGER_LINKAGE   = 4,
        PM_GET_PROTOCOL_INI_PATH          = 5,
        PM_REGISTER_PROTOCOL_MANAGER_INFO = 6,
        PM_INIT_AND_REGISTER              = 7,
        PM_UNBIND_AND_STOP                = 8,
        PM_BIND_STATUS                    = 9,
        PM_REGISTER_STATUS                = 10
      } NdisProtManager;


typedef enum {
        ERR_SUCCESS                      = 0x00,
        ERR_WAIT_FOR_RELEASE             = 0x01,
        ERR_REQUEST_QUEUED               = 0x02,
        ERR_FRAME_NOT_RECOGNIZED         = 0x03,
        ERR_FRAME_REJECTED               = 0x04,
        ERR_FORWARD_FRAME                = 0x05,
        ERR_OUT_OF_RESOURCE              = 0x06,
        ERR_INVALID_PARAMETER            = 0x07,
        ERR_INVALID_FUNCTION             = 0x08,
        ERR_NOT_SUPPORTED                = 0x09,
        ERR_HARDWARE_ERROR               = 0x0A,
        ERR_TRANSMIT_ERROR               = 0x0B,
        ERR_NO_SUCH_DESTINATION          = 0x0C,
        ERR_BUFFER_TOO_SMALL             = 0x0D,
        ERR_ALREADY_STARTED              = 0x20,
        ERR_INCOMPLETE_BINDING           = 0x21,
        ERR_DRIVER_NOT_INITIALIZED       = 0x22,
        ERR_HARDWARE_NOT_FOUND           = 0x23,
        ERR_HARDWARE_FAILURE             = 0x24,
        ERR_CONFIGURATION_FAILURE        = 0x25,
        ERR_INTERRUPT_CONFLICT           = 0x26,
        ERR_INCOMPATIBLE_MAC             = 0x27,
        ERR_INITIALIZATION_FAILED        = 0x28,
        ERR_NO_BINDING                   = 0x29,
        ERR_NETWORK_MAY_NOT_BE_CONNECTED = 0x2A,
        ERR_INCOMPATIBLE_OS_VERSION      = 0x2B,
        ERR_ALREADY_REGISTERED           = 0x2C,
        ERR_PATH_NOT_FOUND               = 0x2D,
        ERR_INSUFFICIENT_MEMORY          = 0x2E,
        ERR_INFO_NOT_FOUND               = 0x2F,
        ERR_GENERAL_FAILURE              = 0xFF
      } NdisError;

#define NDIS_PARAM_INTEGER   0
#define NDIS_PARAM_STRING    1

#define NDIS_TX_BUF_LENGTH   8
#define NDIS_TD_BUF_LENGTH   1
#define NDIS_RX_BUF_LENGTH   8

#define NDIS_PTR_PHYSICAL    0
#define NDIS_PTR_VIRTUAL     2

#define NDIS_PATH    "PROTMAN$"


typedef struct _CommonChars {
        WORD  tableSize;
        BYTE  majorNdisVersion;        /* 2 - Latest version */
        BYTE  minorNdisVersion;        /* 0                  */
        WORD  reserved1;
        BYTE  majorModuleVersion;
        BYTE  minorModuleVersion;
        DWORD moduleFlags;
        /* 0 - Binding at upper boundary supported
         * 1 - Binding at lower boundary supported
         * 2 - Dynamically bound.
         * 3-31 - Reserved, must be zero.
         */
        BYTE  moduleName[16];
        BYTE  protocolLevelUpper;
        /* 1 - MAC
         * 2 - Data Link
         * 3 - Network
         * 4 - Transport
         * 5 - Session
         * -1 - Not specified
         */
        BYTE  interfaceUpper;
        BYTE  protocolLevelLower;
        /* 0 - Physical
         * 1 - MAC
         * 2 - Data Link
         * 3 - Network
         * 4 - Transport
         * 5 - Session
         * -1 - Not specified
         */
        BYTE  interfaceLower;
        WORD  moduleId;
        WORD  moduleDS;
        SystemRequest systemRequest;
        BYTE *serviceChars;
        BYTE *serviceStatus;
        BYTE *upperDispatchTable;
        BYTE *lowerDispatchTable;
        BYTE *reserved2;            /* Must be NULL */
        BYTE *reserved3;            /* Must be NULL */
      } CommonChars;


typedef struct _MulticastList {
        WORD   maxMulticastAddresses;
        WORD   numberMulticastAddresses;
        BYTE   multicastAddress[16][16];
      } MulticastList;


typedef struct _MacChars {
        WORD   tableSize;
        BYTE   macName[16];
        WORD   addressLength;
        BYTE   permanentAddress[16];
        BYTE   currentAddress[16];
        DWORD  currentFunctionalAddress;
        MulticastList *multicastList;
        DWORD  linkSpeed;
        DWORD  serviceFlags;
        WORD   maxFrameSize;
        DWORD  txBufferSize;
        WORD   txBufferAllocSize;
        DWORD  rxBufferSize;
        WORD   rxBufferAllocSize;
        BYTE   ieeeVendor[3];
        BYTE   vendorAdapter;
        BYTE  *vendorAdapterDescription;
        WORD   interruptLevel;
        WORD   txQueueDepth;
        WORD   maxDataBlocks;
      } MacChars;


typedef struct _ProtocolChars {
        WORD   length;
        BYTE   name[16];
        WORD   type;
      } ProtocolChars;


typedef struct _MacUpperDispatch {
        CommonChars      *backPointer;
        Request           request;
        TransmitChain     transmitChain;
        TransferData      transferData;
        ReceiveRelease    receiveRelease;
        IndicationOn      indicationOn;
        IndicationOff     indicationOff;
      } MacUpperDispatch;


typedef struct _MacStatusTable {
        WORD   tableSize;
        DWORD  lastDiag;
        DWORD  macStatus;
        WORD   packetFilter;
        BYTE  *mediaSpecificStats;
        DWORD  lastClear;
        DWORD  totalFramesRx;
        DWORD  totalFramesCrc;
        DWORD  totalBytesRx;
        DWORD  totalDiscardBufSpaceRx;
        DWORD  totalMulticastRx;
        DWORD  totalBroadcastRx;
        DWORD  obsolete1[5];
        DWORD  totalDiscardHwErrorRx;
        DWORD  totalFramesTx;
        DWORD  totalBytesTx;
        DWORD  totalMulticastTx;
        DWORD  totalBroadcastTx;
        DWORD  obsolete2[2];
        DWORD  totalDiscardTimeoutTx;
        DWORD  totalDiscardHwErrorTx;
      } MacStatusTable;


typedef struct _ProtDispatch {
        CommonChars *backPointer;
        DWORD        flags;
        /* 0 - handles non-LLC frames
         * 1 - handles specific-LSAP LLC frames
         * 2 - handles specific-LSAP LLC frames
         * 3-31 - reserved must be 0
         */
        void  (*requestConfirm) (void);
        void  (*transmitConfirm) (void);
        void  (*receiveLookahead) (void);
        void  (*indicationComplete) (void);
        void  (*receiveChain) (void);
        void  (*status) (void);
      } ProtDispatch;


typedef struct _ReqBlock {
        WORD      opcode;
        WORD      status;
        BYTE FAR *pointer1;
        BYTE FAR *pointer2;
        WORD      word1;
      } ReqBlock;


typedef struct _TxBufDescrRec {
        BYTE   txPtrType;
        BYTE   dummy;
        WORD   txDataLen;
        BYTE  *txDataPtr;
      } TxBufDescrRec;


typedef struct _TxBufDescr {
        WORD          txImmedLen;
        BYTE         *txImmedPtr;
        WORD          txDataCount;
        TxBufDescrRec txBufDescrRec[NDIS_TX_BUF_LENGTH];
      } TxBufDescr;


typedef struct _TDBufDescrRec {
        BYTE   tDPtrType;
        BYTE   dummy;
        WORD   tDDataLen;
        BYTE  *tDDataPtr;
      } TDBufDescrRec;


typedef struct _TDBufDescr {
        WORD          tDDataCount;
        TDBufDescrRec tDBufDescrRec[NDIS_TD_BUF_LENGTH];
      } TDBufDescr;


typedef struct _RxBufDescrRec {
        WORD   rxDataLen;
        BYTE  *rxDataPtr;
      } RxBufDescrRec;


typedef struct _RxBufDescr {
        WORD          rxDataCount;
        RxBufDescrRec rxBufDescrRec[NDIS_RX_BUF_LENGTH];
      } RxBufDescr;


typedef struct _PktBuf {
	struct _PktBuf *nextLink;
	struct _PktBuf *prevLink;
        int    handle;
        int    length;
        int    packetLength;
        DWORD  sequence;
        BYTE  *buffer;
      } PktBuf;


typedef struct _CardHandle {
        BYTE         moduleName[16];
        CommonChars *common;
      } CardHandle;


typedef struct _BindingsList {
        WORD  numBindings;
        BYTE  moduleName[2][16];
      } BindingsList;


typedef struct _FailingModules {
        BYTE  upperModuleName[16];
        BYTE  lowerModuleName[16];
      } FailingModules;


typedef union _HardwareAddress {
        BYTE  bytes[6];
        WORD  words[3];
        struct {
          BYTE bytes[6];
        } addr;
      } HardwareAddress;


typedef struct _FddiHeader {
        BYTE             frameControl;
        HardwareAddress  etherDestHost;
        HardwareAddress  etherSrcHost;
      } FddiHeader;


typedef struct _EthernetIIHeader {
        HardwareAddress  etherDestHost;
        HardwareAddress  etherSrcHost;
        WORD             etherType;
      } EthernetIIHeader;


typedef struct _Ieee802Dot5Header {
        HardwareAddress  etherDestHost;
        HardwareAddress  etherSrcHost;
        BYTE             routeInfo[30];
      } Ieee802Dot5Header;


typedef struct _Ieee802Dot2SnapHeader {
        BYTE  dsap;                      /* 0xAA */
        BYTE  ssap;                      /* 0xAA */
        BYTE  control;                   /* 3    */
        BYTE protocolId[5];
      } Ieee802Dot2SnapHeader;


/*
 *  Prototypes
 */
extern char *NdisLastError        (void);
extern int   NdisOpen             (void);
extern int   NdisInit             (int promis);
extern int   NdisRegisterAndBind  (int promis);
extern void  NdisShutdown         (void);
extern void  NdisCheckMacFeatures (struct _CardHandle *card);
extern int   NdisSendPacket       (struct _PktBuf *pktBuf, int macId);

/*
 *  Assembly "glue" functions
 */
extern int systemRequestGlue();
extern int requestConfirmGlue();
extern int transmitConfirmGlue();
extern int receiveLookaheadGlue();
extern int indicationCompleteGlue();
extern int receiveChainGlue();
extern int statusGlue();

/*
 *  IOCTL function
 */
#ifdef __SMALL__
extern int _far NdisGetLinkage (int handle, char *data, int size);
#else
extern int NdisGetLinkage (int handle, char *data, int size);
#endif

/*
 *  NDIS callback handlers
 */
CALLBACK (NdisSystemRequest     (DWORD,DWORD, WORD, WORD, WORD));
CALLBACK (NdisRequestConfirm    ( WORD, WORD, WORD, WORD, WORD,WORD));
CALLBACK (NdisTransmitConfirm   ( WORD, WORD, WORD, WORD, WORD));
CALLBACK (NdisReceiveLookahead  ( WORD, WORD, WORD, BYTE*, BYTE*, WORD));
CALLBACK (NdisReceiveChain      ( WORD, WORD, WORD, struct _RxBufDescr*, BYTE*, WORD));
CALLBACK (NdisStatusProc        ( WORD, WORD, BYTE*, WORD,WORD));
CALLBACK (NdisIndicationComplete( WORD, WORD));

BYTE *NdisAllocStack (void);
void  NdisFreeStack  (BYTE*);

#ifdef __HIGHC__
  #define RENAME_ASM_SYM(x) pragma Alias(x,"@" #x "")  /* prepend `@' */
  #define RENAME_C_SYM(x)   pragma Alias(x,"_" #x "")  /* prepend `_' */

  RENAME_ASM_SYM (systemRequestGlue);
  RENAME_ASM_SYM (requestConfirmGlue);
  RENAME_ASM_SYM (transmitConfirmGlue);
  RENAME_ASM_SYM (receiveLookaheadGlue);
  RENAME_ASM_SYM (indicationCompleteGlue);
  RENAME_ASM_SYM (receiveChainGlue);
  RENAME_ASM_SYM (statusGlue);
  RENAME_ASM_SYM (NdisGetLinkage);
  RENAME_C_SYM   (NdisSystemRequest);
  RENAME_C_SYM   (NdisRequestConfirm);
  RENAME_C_SYM   (NdisTransmitConfirm);
  RENAME_C_SYM   (NdisReceiveLookahead);
  RENAME_C_SYM   (NdisIndicationComplete);
  RENAME_C_SYM   (NdisReceiveChain);
  RENAME_C_SYM   (NdisStatusProc);
  RENAME_C_SYM   (NdisAllocStack);
  RENAME_C_SYM   (NdisFreeStack);
#endif

#endif
