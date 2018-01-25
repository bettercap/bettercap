/*-
 * Copyright (c) 1990, 1991, 1992, 1993, 1994, 1995, 1996, 1997
 *	The Regents of the University of California.  All rights reserved.
 *
 * This code is derived from the Stanford/CMU enet packet filter,
 * (net/enet.c) distributed as part of 4.3BSD, and code contributed
 * to Berkeley by Steven McCanne and Van Jacobson both of Lawrence
 * Berkeley Laboratory.
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
 *	This product includes software developed by the University of
 *	California, Berkeley and its contributors.
 * 4. Neither the name of the University nor the names of its contributors
 *    may be used to endorse or promote products derived from this software
 *    without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS ``AS IS'' AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
 * OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 *
 *	@(#)bpf.c	7.5 (Berkeley) 7/15/91
 */

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#ifdef _WIN32

#include <pcap-stdinc.h>

#else /* _WIN32 */

#if HAVE_INTTYPES_H
#include <inttypes.h>
#elif HAVE_STDINT_H
#include <stdint.h>
#endif
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif

#include <sys/param.h>
#include <sys/types.h>
#include <sys/time.h>

#define	SOLARIS	(defined(sun) && (defined(__SVR4) || defined(__svr4__)))
#if defined(__hpux) || SOLARIS
# include <sys/sysmacros.h>
# include <sys/stream.h>
# define	mbuf	msgb
# define	m_next	b_cont
# define	MLEN(m)	((m)->b_wptr - (m)->b_rptr)
# define	mtod(m,t)	((t)(m)->b_rptr)
#else /* defined(__hpux) || SOLARIS */
# define	MLEN(m)	((m)->m_len)
#endif /* defined(__hpux) || SOLARIS */

#endif /* _WIN32 */

#include <pcap/bpf.h>

#if !defined(KERNEL) && !defined(_KERNEL)
#include <stdlib.h>
#endif

#define int32 bpf_int32
#define u_int32 bpf_u_int32

#ifndef LBL_ALIGN
/*
 * XXX - IA-64?  If not, this probably won't work on Win64 IA-64
 * systems, unless LBL_ALIGN is defined elsewhere for them.
 * XXX - SuperH?  If not, this probably won't work on WinCE SuperH
 * systems, unless LBL_ALIGN is defined elsewhere for them.
 */
#if defined(sparc) || defined(__sparc__) || defined(mips) || \
    defined(ibm032) || defined(__alpha) || defined(__hpux) || \
    defined(__arm__)
#define LBL_ALIGN
#endif
#endif

#ifndef LBL_ALIGN
#ifndef _WIN32
#include <netinet/in.h>
#endif

#define EXTRACT_SHORT(p)	((u_short)ntohs(*(u_short *)p))
#define EXTRACT_LONG(p)		(ntohl(*(u_int32 *)p))
#else
#define EXTRACT_SHORT(p)\
	((u_short)\
		((u_short)*((u_char *)p+0)<<8|\
		 (u_short)*((u_char *)p+1)<<0))
#define EXTRACT_LONG(p)\
		((u_int32)*((u_char *)p+0)<<24|\
		 (u_int32)*((u_char *)p+1)<<16|\
		 (u_int32)*((u_char *)p+2)<<8|\
		 (u_int32)*((u_char *)p+3)<<0)
#endif

#if defined(KERNEL) || defined(_KERNEL)
# if !defined(__hpux) && !SOLARIS
#include <sys/mbuf.h>
# endif
#define MINDEX(len, _m, _k) \
{ \
	len = MLEN(m); \
	while ((_k) >= len) { \
		(_k) -= len; \
		(_m) = (_m)->m_next; \
		if ((_m) == 0) \
			return 0; \
		len = MLEN(m); \
	} \
}

static int
m_xword(m, k, err)
	register struct mbuf *m;
	register int k, *err;
{
	register int len;
	register u_char *cp, *np;
	register struct mbuf *m0;

	MINDEX(len, m, k);
	cp = mtod(m, u_char *) + k;
	if (len - k >= 4) {
		*err = 0;
		return EXTRACT_LONG(cp);
	}
	m0 = m->m_next;
	if (m0 == 0 || MLEN(m0) + len - k < 4)
		goto bad;
	*err = 0;
	np = mtod(m0, u_char *);
	switch (len - k) {

	case 1:
		return (cp[0] << 24) | (np[0] << 16) | (np[1] << 8) | np[2];

	case 2:
		return (cp[0] << 24) | (cp[1] << 16) | (np[0] << 8) | np[1];

	default:
		return (cp[0] << 24) | (cp[1] << 16) | (cp[2] << 8) | np[0];
	}
    bad:
	*err = 1;
	return 0;
}

static int
m_xhalf(m, k, err)
	register struct mbuf *m;
	register int k, *err;
{
	register int len;
	register u_char *cp;
	register struct mbuf *m0;

	MINDEX(len, m, k);
	cp = mtod(m, u_char *) + k;
	if (len - k >= 2) {
		*err = 0;
		return EXTRACT_SHORT(cp);
	}
	m0 = m->m_next;
	if (m0 == 0)
		goto bad;
	*err = 0;
	return (cp[0] << 8) | mtod(m0, u_char *)[0];
 bad:
	*err = 1;
	return 0;
}
#endif

#ifdef __linux__
#include <linux/types.h>
#include <linux/if_packet.h>
#include <linux/filter.h>
#endif

enum {
        BPF_S_ANC_NONE,
        BPF_S_ANC_VLAN_TAG,
        BPF_S_ANC_VLAN_TAG_PRESENT,
};

/*
 * Execute the filter program starting at pc on the packet p
 * wirelen is the length of the original packet
 * buflen is the amount of data present
 * aux_data is auxiliary data, currently used only when interpreting
 * filters intended for the Linux kernel in cases where the kernel
 * rejects the filter; it contains VLAN tag information
 * For the kernel, p is assumed to be a pointer to an mbuf if buflen is 0,
 * in all other cases, p is a pointer to a buffer and buflen is its size.
 *
 * Thanks to Ani Sinha <ani@arista.com> for providing initial implementation
 */
u_int
bpf_filter_with_aux_data(pc, p, wirelen, buflen, aux_data)
	register const struct bpf_insn *pc;
	register const u_char *p;
	u_int wirelen;
	register u_int buflen;
	register const struct bpf_aux_data *aux_data;
{
	register u_int32 A, X;
	register bpf_u_int32 k;
	u_int32 mem[BPF_MEMWORDS];
#if defined(KERNEL) || defined(_KERNEL)
	struct mbuf *m, *n;
	int merr, len;

	if (buflen == 0) {
		m = (struct mbuf *)p;
		p = mtod(m, u_char *);
		buflen = MLEN(m);
	} else
		m = NULL;
#endif

	if (pc == 0)
		/*
		 * No filter means accept all.
		 */
		return (u_int)-1;
	A = 0;
	X = 0;
	--pc;
	while (1) {
		++pc;
		switch (pc->code) {

		default:
#if defined(KERNEL) || defined(_KERNEL)
			return 0;
#else
			abort();
#endif
		case BPF_RET|BPF_K:
			return (u_int)pc->k;

		case BPF_RET|BPF_A:
			return (u_int)A;

		case BPF_LD|BPF_W|BPF_ABS:
			k = pc->k;
			if (k > buflen || sizeof(int32_t) > buflen - k) {
#if defined(KERNEL) || defined(_KERNEL)
				if (m == NULL)
					return 0;
				A = m_xword(m, k, &merr);
				if (merr != 0)
					return 0;
				continue;
#else
				return 0;
#endif
			}
			A = EXTRACT_LONG(&p[k]);
			continue;

		case BPF_LD|BPF_H|BPF_ABS:
			k = pc->k;
			if (k > buflen || sizeof(int16_t) > buflen - k) {
#if defined(KERNEL) || defined(_KERNEL)
				if (m == NULL)
					return 0;
				A = m_xhalf(m, k, &merr);
				if (merr != 0)
					return 0;
				continue;
#else
				return 0;
#endif
			}
			A = EXTRACT_SHORT(&p[k]);
			continue;

		case BPF_LD|BPF_B|BPF_ABS:
			{
#if defined(SKF_AD_VLAN_TAG) && defined(SKF_AD_VLAN_TAG_PRESENT)
				int code = BPF_S_ANC_NONE;
#define ANCILLARY(CODE) case SKF_AD_OFF + SKF_AD_##CODE:		\
				code = BPF_S_ANC_##CODE;		\
                                        if (!aux_data)                  \
                                                return 0;               \
                                        break;

				switch (pc->k) {
					ANCILLARY(VLAN_TAG);
					ANCILLARY(VLAN_TAG_PRESENT);
				default :
#endif
					k = pc->k;
					if (k >= buflen) {
#if defined(KERNEL) || defined(_KERNEL)
						if (m == NULL)
							return 0;
						n = m;
						MINDEX(len, n, k);
						A = mtod(n, u_char *)[k];
						continue;
#else
						return 0;
#endif
					}
					A = p[k];
#if defined(SKF_AD_VLAN_TAG) && defined(SKF_AD_VLAN_TAG_PRESENT)
				}
				switch (code) {
				case BPF_S_ANC_VLAN_TAG:
					if (aux_data)
						A = aux_data->vlan_tag;
					break;

				case BPF_S_ANC_VLAN_TAG_PRESENT:
					if (aux_data)
						A = aux_data->vlan_tag_present;
					break;
				}
#endif
				continue;
			}
		case BPF_LD|BPF_W|BPF_LEN:
			A = wirelen;
			continue;

		case BPF_LDX|BPF_W|BPF_LEN:
			X = wirelen;
			continue;

		case BPF_LD|BPF_W|BPF_IND:
			k = X + pc->k;
			if (pc->k > buflen || X > buflen - pc->k ||
			    sizeof(int32_t) > buflen - k) {
#if defined(KERNEL) || defined(_KERNEL)
				if (m == NULL)
					return 0;
				A = m_xword(m, k, &merr);
				if (merr != 0)
					return 0;
				continue;
#else
				return 0;
#endif
			}
			A = EXTRACT_LONG(&p[k]);
			continue;

		case BPF_LD|BPF_H|BPF_IND:
			k = X + pc->k;
			if (X > buflen || pc->k > buflen - X ||
			    sizeof(int16_t) > buflen - k) {
#if defined(KERNEL) || defined(_KERNEL)
				if (m == NULL)
					return 0;
				A = m_xhalf(m, k, &merr);
				if (merr != 0)
					return 0;
				continue;
#else
				return 0;
#endif
			}
			A = EXTRACT_SHORT(&p[k]);
			continue;

		case BPF_LD|BPF_B|BPF_IND:
			k = X + pc->k;
			if (pc->k >= buflen || X >= buflen - pc->k) {
#if defined(KERNEL) || defined(_KERNEL)
				if (m == NULL)
					return 0;
				n = m;
				MINDEX(len, n, k);
				A = mtod(n, u_char *)[k];
				continue;
#else
				return 0;
#endif
			}
			A = p[k];
			continue;

		case BPF_LDX|BPF_MSH|BPF_B:
			k = pc->k;
			if (k >= buflen) {
#if defined(KERNEL) || defined(_KERNEL)
				if (m == NULL)
					return 0;
				n = m;
				MINDEX(len, n, k);
				X = (mtod(n, char *)[k] & 0xf) << 2;
				continue;
#else
				return 0;
#endif
			}
			X = (p[pc->k] & 0xf) << 2;
			continue;

		case BPF_LD|BPF_IMM:
			A = pc->k;
			continue;

		case BPF_LDX|BPF_IMM:
			X = pc->k;
			continue;

		case BPF_LD|BPF_MEM:
			A = mem[pc->k];
			continue;

		case BPF_LDX|BPF_MEM:
			X = mem[pc->k];
			continue;

		case BPF_ST:
			mem[pc->k] = A;
			continue;

		case BPF_STX:
			mem[pc->k] = X;
			continue;

		case BPF_JMP|BPF_JA:
#if defined(KERNEL) || defined(_KERNEL)
			/*
			 * No backward jumps allowed.
			 */
			pc += pc->k;
#else
			/*
			 * XXX - we currently implement "ip6 protochain"
			 * with backward jumps, so sign-extend pc->k.
			 */
			pc += (bpf_int32)pc->k;
#endif
			continue;

		case BPF_JMP|BPF_JGT|BPF_K:
			pc += (A > pc->k) ? pc->jt : pc->jf;
			continue;

		case BPF_JMP|BPF_JGE|BPF_K:
			pc += (A >= pc->k) ? pc->jt : pc->jf;
			continue;

		case BPF_JMP|BPF_JEQ|BPF_K:
			pc += (A == pc->k) ? pc->jt : pc->jf;
			continue;

		case BPF_JMP|BPF_JSET|BPF_K:
			pc += (A & pc->k) ? pc->jt : pc->jf;
			continue;

		case BPF_JMP|BPF_JGT|BPF_X:
			pc += (A > X) ? pc->jt : pc->jf;
			continue;

		case BPF_JMP|BPF_JGE|BPF_X:
			pc += (A >= X) ? pc->jt : pc->jf;
			continue;

		case BPF_JMP|BPF_JEQ|BPF_X:
			pc += (A == X) ? pc->jt : pc->jf;
			continue;

		case BPF_JMP|BPF_JSET|BPF_X:
			pc += (A & X) ? pc->jt : pc->jf;
			continue;

		case BPF_ALU|BPF_ADD|BPF_X:
			A += X;
			continue;

		case BPF_ALU|BPF_SUB|BPF_X:
			A -= X;
			continue;

		case BPF_ALU|BPF_MUL|BPF_X:
			A *= X;
			continue;

		case BPF_ALU|BPF_DIV|BPF_X:
			if (X == 0)
				return 0;
			A /= X;
			continue;

		case BPF_ALU|BPF_MOD|BPF_X:
			if (X == 0)
				return 0;
			A %= X;
			continue;

		case BPF_ALU|BPF_AND|BPF_X:
			A &= X;
			continue;

		case BPF_ALU|BPF_OR|BPF_X:
			A |= X;
			continue;

		case BPF_ALU|BPF_XOR|BPF_X:
			A ^= X;
			continue;

		case BPF_ALU|BPF_LSH|BPF_X:
			A <<= X;
			continue;

		case BPF_ALU|BPF_RSH|BPF_X:
			A >>= X;
			continue;

		case BPF_ALU|BPF_ADD|BPF_K:
			A += pc->k;
			continue;

		case BPF_ALU|BPF_SUB|BPF_K:
			A -= pc->k;
			continue;

		case BPF_ALU|BPF_MUL|BPF_K:
			A *= pc->k;
			continue;

		case BPF_ALU|BPF_DIV|BPF_K:
			A /= pc->k;
			continue;

		case BPF_ALU|BPF_MOD|BPF_K:
			A %= pc->k;
			continue;

		case BPF_ALU|BPF_AND|BPF_K:
			A &= pc->k;
			continue;

		case BPF_ALU|BPF_OR|BPF_K:
			A |= pc->k;
			continue;

		case BPF_ALU|BPF_XOR|BPF_K:
			A ^= pc->k;
			continue;

		case BPF_ALU|BPF_LSH|BPF_K:
			A <<= pc->k;
			continue;

		case BPF_ALU|BPF_RSH|BPF_K:
			A >>= pc->k;
			continue;

		case BPF_ALU|BPF_NEG:
			/*
			 * Most BPF arithmetic is unsigned, but negation
			 * can't be unsigned; throw some casts to
			 * specify what we're trying to do.
			 */
			A = (u_int32)(-(int32)A);
			continue;

		case BPF_MISC|BPF_TAX:
			X = A;
			continue;

		case BPF_MISC|BPF_TXA:
			A = X;
			continue;
		}
	}
}

u_int
bpf_filter(pc, p, wirelen, buflen)
	register const struct bpf_insn *pc;
	register const u_char *p;
	u_int wirelen;
	register u_int buflen;
{
	return bpf_filter_with_aux_data(pc, p, wirelen, buflen, NULL);
}


/*
 * Return true if the 'fcode' is a valid filter program.
 * The constraints are that each jump be forward and to a valid
 * code, that memory accesses are within valid ranges (to the
 * extent that this can be checked statically; loads of packet
 * data have to be, and are, also checked at run time), and that
 * the code terminates with either an accept or reject.
 *
 * The kernel needs to be able to verify an application's filter code.
 * Otherwise, a bogus program could easily crash the system.
 */
int
bpf_validate(f, len)
	const struct bpf_insn *f;
	int len;
{
	u_int i, from;
	const struct bpf_insn *p;

	if (len < 1)
		return 0;
	/*
	 * There's no maximum program length in userland.
	 */
#if defined(KERNEL) || defined(_KERNEL)
	if (len > BPF_MAXINSNS)
		return 0;
#endif

	for (i = 0; i < (u_int)len; ++i) {
		p = &f[i];
		switch (BPF_CLASS(p->code)) {
		/*
		 * Check that memory operations use valid addresses.
		 */
		case BPF_LD:
		case BPF_LDX:
			switch (BPF_MODE(p->code)) {
			case BPF_IMM:
				break;
			case BPF_ABS:
			case BPF_IND:
			case BPF_MSH:
				/*
				 * There's no maximum packet data size
				 * in userland.  The runtime packet length
				 * check suffices.
				 */
#if defined(KERNEL) || defined(_KERNEL)
				/*
				 * More strict check with actual packet length
				 * is done runtime.
				 */
				if (p->k >= bpf_maxbufsize)
					return 0;
#endif
				break;
			case BPF_MEM:
				if (p->k >= BPF_MEMWORDS)
					return 0;
				break;
			case BPF_LEN:
				break;
			default:
				return 0;
			}
			break;
		case BPF_ST:
		case BPF_STX:
			if (p->k >= BPF_MEMWORDS)
				return 0;
			break;
		case BPF_ALU:
			switch (BPF_OP(p->code)) {
			case BPF_ADD:
			case BPF_SUB:
			case BPF_MUL:
			case BPF_OR:
			case BPF_AND:
			case BPF_XOR:
			case BPF_LSH:
			case BPF_RSH:
			case BPF_NEG:
				break;
			case BPF_DIV:
			case BPF_MOD:
				/*
				 * Check for constant division or modulus
				 * by 0.
				 */
				if (BPF_SRC(p->code) == BPF_K && p->k == 0)
					return 0;
				break;
			default:
				return 0;
			}
			break;
		case BPF_JMP:
			/*
			 * Check that jumps are within the code block,
			 * and that unconditional branches don't go
			 * backwards as a result of an overflow.
			 * Unconditional branches have a 32-bit offset,
			 * so they could overflow; we check to make
			 * sure they don't.  Conditional branches have
			 * an 8-bit offset, and the from address is <=
			 * BPF_MAXINSNS, and we assume that BPF_MAXINSNS
			 * is sufficiently small that adding 255 to it
			 * won't overflow.
			 *
			 * We know that len is <= BPF_MAXINSNS, and we
			 * assume that BPF_MAXINSNS is < the maximum size
			 * of a u_int, so that i + 1 doesn't overflow.
			 *
			 * For userland, we don't know that the from
			 * or len are <= BPF_MAXINSNS, but we know that
			 * from <= len, and, except on a 64-bit system,
			 * it's unlikely that len, if it truly reflects
			 * the size of the program we've been handed,
			 * will be anywhere near the maximum size of
			 * a u_int.  We also don't check for backward
			 * branches, as we currently support them in
			 * userland for the protochain operation.
			 */
			from = i + 1;
			switch (BPF_OP(p->code)) {
			case BPF_JA:
#if defined(KERNEL) || defined(_KERNEL)
				if (from + p->k < from || from + p->k >= len)
#else
				if (from + p->k >= (u_int)len)
#endif
					return 0;
				break;
			case BPF_JEQ:
			case BPF_JGT:
			case BPF_JGE:
			case BPF_JSET:
				if (from + p->jt >= (u_int)len || from + p->jf >= (u_int)len)
					return 0;
				break;
			default:
				return 0;
			}
			break;
		case BPF_RET:
			break;
		case BPF_MISC:
			break;
		default:
			return 0;
		}
	}
	return BPF_CLASS(f[len - 1].code) == BPF_RET;
}
