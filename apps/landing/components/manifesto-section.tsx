'use client';

import { motion, useInView } from 'motion/react';
import { useRef } from 'react';

const principles = [
  'Encrypt in the repo. Decrypt in the editor.',
  'One master key. Every secret wrapped.',
  "Ciphertext travels. The key doesn't.",
];

export function ManifestoSection() {
  const ref = useRef<HTMLElement>(null);
  const isInView = useInView(ref, { amount: 0.3, once: true });

  return (
    <section
      aria-labelledby="manifesto-heading"
      className="relative overflow-hidden py-20 md:py-28"
      ref={ref}
    >
      <div aria-hidden className="absolute inset-0 bg-black" />
      <div
        aria-hidden
        className="absolute inset-0 opacity-[0.03]"
        style={{
          backgroundImage: `url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E")`,
        }}
      />

      <div className="relative mx-auto max-w-4xl px-6 text-center">
        <motion.p
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          className="font-medium text-2xl text-white leading-snug tracking-tight md:text-3xl lg:text-4xl"
          id="manifesto-heading"
          initial={{ opacity: 0, y: 16 }}
          transition={{
            duration: 0.55,
            ease: [0.25, 0.46, 0.45, 0.94],
          }}
        >
          If it’s not encrypted,
          <br />
          <span className="text-white/70">it’s not a secret.</span>
        </motion.p>

        <motion.p
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          className="mx-auto mt-6 max-w-2xl font-mono text-[13px] text-white/50 leading-relaxed tracking-tight md:text-sm"
          initial={{ opacity: 0, y: 12 }}
          transition={{
            duration: 0.5,
            delay: 0.15,
            ease: [0.25, 0.46, 0.45, 0.94],
          }}
        >
          In the repo, in CI, in every handoff-ciphertext only. OpenEnvX keeps
          it that way.
        </motion.p>

        <ul className="mx-auto mt-10 flex max-w-2xl flex-col items-center gap-4 text-center font-mono text-[13px] text-white/40 md:mt-12 md:gap-5 md:text-sm">
          {principles.map((line, i) => (
            <motion.li
              animate={isInView ? { opacity: 1, x: 0 } : {}}
              className="flex items-center justify-center gap-3 before:inline-block before:h-px before:w-6 before:shrink-0"
              initial={{ opacity: 0, x: -8 }}
              key={line}
              transition={{
                duration: 0.45,
                delay: 0.25 + i * 0.08,
                ease: [0.25, 0.46, 0.45, 0.94],
              }}
            >
              {line}
            </motion.li>
          ))}
        </ul>
      </div>
    </section>
  );
}
