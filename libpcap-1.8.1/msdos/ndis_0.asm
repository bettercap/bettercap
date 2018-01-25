PAGE 60,132
NAME NDIS_0

ifdef DOSX
  .386
  _TEXT   SEGMENT PUBLIC DWORD USE16 'CODE'
  _TEXT   ENDS
  _DATA   SEGMENT PUBLIC DWORD USE16 'CODE'
  _DATA   ENDS
  _TEXT32 SEGMENT PUBLIC BYTE  USE32 'CODE'
  _TEXT32 ENDS
  CB_DSEG EQU <CS>                          ; DOSX is tiny-model
  D_SEG   EQU <_TEXT SEGMENT>
  D_END   EQU <_TEXT ENDS>
  ASSUME  CS:_TEXT,DS:_TEXT

  PUSHREGS equ <pushad>
  POPREGS  equ <popad>

  PUBPROC macro name
          align 4
          public @&name
          @&name label near
          endm
else
  .286
  _TEXT   SEGMENT PUBLIC DWORD 'CODE'
  _TEXT   ENDS
  _DATA   SEGMENT PUBLIC DWORD 'DATA'
  _DATA   ENDS
  CB_DSEG EQU <SEG _DATA>                   ; 16bit is small/large model
  D_SEG   EQU <_DATA SEGMENT>
  D_END   EQU <_DATA ENDS>
  ASSUME  CS:_TEXT,DS:_DATA

  PUSHREGS equ <pusha>
  POPREGS  equ <popa>

  PUBPROC  macro name
           public _&name
           _&name label far
           endm
endif

;-------------------------------------------

D_SEG

D_END


_TEXT SEGMENT

EXTRN _NdisSystemRequest      : near
EXTRN _NdisRequestConfirm     : near
EXTRN _NdisTransmitConfirm    : near
EXTRN _NdisReceiveLookahead   : near
EXTRN _NdisIndicationComplete : near
EXTRN _NdisReceiveChain       : near
EXTRN _NdisStatusProc         : near
EXTRN _NdisAllocStack         : near
EXTRN _NdisFreeStack          : near

;
; *ALL* interrupt threads come through this macro.
;
CALLBACK macro callbackProc, argsSize

     pushf
     PUSHREGS                ;; Save the registers

     push es
     push ds
     mov  ax,CB_DSEG         ;; Load DS
     mov  ds,ax
     call _NdisAllocStack    ;; Get and install a stack.

     mov  bx,ss              ;; Save off the old stack in other regs
     mov  cx,sp
     mov  ss,dx              ;; Install the new one
     mov  sp,ax
     push bx                 ;; Save the old one on to the new stack
     push cx
     sub  sp,&argsSize       ;; Allocate space for arguments on the stack

     mov  ax,ss              ;; Set up the destination for the move
     mov  es,ax
     mov  di,sp
     mov  ds,bx              ;; Set up the source for the move.
     mov  si,cx
     add  si,4+6+32

     mov  cx,&argsSize       ;; Move the arguments to the stack.
     shr  cx,1
     cld
     rep  movsw

     mov  ax,CB_DSEG         ;; Set my data segment again.
     mov  ds,ax

     call &callbackProc      ;; Call the real callback.
     pop  di                 ;; Pop off the old stack
     pop  si
     mov  bx,ss              ;; Save off the current allocated stack.
     mov  cx,sp
     mov  ss,si              ;; Restore the old stack
     mov  sp,di
     push ax                 ;; Save the return code
     push bx                 ;; Free the stack. Push the pointer to it
     push cx
     call _NdisFreeStack
     add  sp,4
     pop  ax                 ;; Get the return code back
     add  di,32              ;; Get a pointer to ax on the stack
     mov  word ptr ss:[di],ax
     pop  ds
     pop  es

     POPREGS
     popf
endm

;
; Define all of the callbacks for the NDIS procs.
;

PUBPROC systemRequestGlue
CALLBACK _NdisSystemRequest,14
RETF

PUBPROC requestConfirmGlue
CALLBACK _NdisRequestConfirm,12
RETF

PUBPROC transmitConfirmGlue
CALLBACK _NdisTransmitConfirm,10
RETF

PUBPROC receiveLookaheadGlue
CALLBACK _NdisReceiveLookahead,16
RETF

PUBPROC indicationCompleteGlue
CALLBACK _NdisIndicationComplete,4
RETF

PUBPROC receiveChainGlue
CALLBACK _NdisReceiveChain,16
RETF

PUBPROC statusGlue
CALLBACK _NdisStatusProc,12
RETF

;
; int FAR NdisGetLinkage (int handle, char *data, int size);
;

ifdef DOSX
  PUBPROC NdisGetLinkage
          push ebx
          mov ebx, [esp+8]              ; device handle
          mov eax, 4402h                ; IOCTRL read function
          mov edx, [esp+12]             ; DS:EDX -> result data
          mov ecx, [esp+16]             ; ECX = length
          int 21h
          pop ebx
          jc  @fail
          xor eax, eax
  @fail:  ret

else
  PUBPROC NdisGetLinkage
          enter 0, 0
          mov bx, [bp+6]
          mov ax, 4402h
          mov dx, [bp+8]
          mov cx, [bp+12]
          int 21h
          jc  @fail
          xor ax, ax
  @fail:  leave
          retf
endif

ENDS

END
