'use client';

import { motion, useInView } from 'motion/react';
import { useRef } from 'react';
import {
  siBun,
  siDeno,
  siDocker,
  siGithub,
  siGitlab,
  siKubernetes,
  siNextdotjs,
  siNodedotjs,
  siNpm,
  siPnpm,
  siVite,
  siYarn,
} from 'simple-icons';

const iconMap = {
  'Next.js': siNextdotjs,
  'Node.js': siNodedotjs,
  Bun: siBun,
  Deno: siDeno,
  Vite: siVite,
  'GitHub Actions': siGithub,
  'GitLab CI': siGitlab,
  npm: siNpm,
  pnpm: siPnpm,
  Yarn: siYarn,
  Docker: siDocker,
  Kubernetes: siKubernetes,
} as const;

const groups = [
  {
    label: 'Runtimes & frameworks',
    items: ['Node.js', 'Bun', 'Deno', 'Next.js', 'Vite'] as const,
    description: 'Run openenvx in any JS/TS runtime or framework.',
  },
  {
    label: 'CI / CD',
    items: ['GitHub Actions', 'GitLab CI'] as const,
    description: 'Decrypt in pipeline. No secrets in logs.',
  },
  {
    label: 'Package managers',
    items: ['npm', 'pnpm', 'Yarn'] as const,
    description: 'Install once, use in scripts and dev.',
  },
  {
    label: 'Containers & orchestration',
    items: ['Docker', 'Kubernetes'] as const,
    description: 'Same binary in containers and clusters.',
  },
];

function BrandIcon({
  slug,
  name,
  size = 28,
}: {
  slug: (typeof iconMap)[keyof typeof iconMap];
  name: string;
  size?: number;
}) {
  const path = slug.path;
  const fill = `#${slug.hex}`;
  return (
    <svg
      aria-hidden
      className="shrink-0"
      fill={fill}
      height={size}
      role="img"
      viewBox="0 0 24 24"
      width={size}
      xmlns="http://www.w3.org/2000/svg"
    >
      <title>{name}</title>
      <path d={path} />
    </svg>
  );
}

export function WorksWithSection() {
  const ref = useRef<HTMLElement>(null);
  const isInView = useInView(ref, { amount: 0.15, once: true });

  return (
    <section
      aria-labelledby="works-with-heading"
      className="section-grain relative overflow-hidden border-[#e5e5e5] border-t py-20 md:py-28"
      ref={ref}
    >
      <div aria-hidden className="absolute inset-0 bg-[#fafafa]/50" />
      <div
        aria-hidden
        className="absolute inset-x-0 top-0 h-px bg-linear-to-r from-transparent via-[#e5e5e5] to-transparent"
      />

      <div className="relative mx-auto max-w-5xl px-6">
        <motion.div
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          className="text-center"
          initial={{ opacity: 0, y: 10 }}
          transition={{ duration: 0.45, ease: [0.25, 0.46, 0.45, 0.94] }}
        >
          <p className="font-mono text-[#737373] text-xs uppercase tracking-widest">
            Compatibility
          </p>
          <h2
            className="mt-2 font-medium text-2xl text-black tracking-tight md:text-3xl"
            id="works-with-heading"
          >
            Works with
          </h2>
        </motion.div>

        <div className="mt-12 grid grid-cols-1 gap-10 sm:grid-cols-2 lg:grid-cols-4 lg:gap-8">
          {groups.map((group, gi) => (
            <motion.div
              animate={isInView ? { opacity: 1, y: 0 } : {}}
              className="flex flex-col rounded-xl border border-[#e5e5e5] bg-white/80 p-5 shadow-[0_1px_2px_rgba(0,0,0,0.04)] transition-shadow hover:shadow-[0_4px_12px_rgba(0,0,0,0.06)]"
              initial={{ opacity: 0, y: 14 }}
              key={group.label}
              transition={{
                duration: 0.4,
                delay: 0.08 + gi * 0.06,
                ease: [0.25, 0.46, 0.45, 0.94],
              }}
            >
              <span className="font-mono text-[#a3a3a3] text-[10px] uppercase tracking-widest">
                {group.label}
              </span>
              <p className="mt-1.5 text-[#525252] text-xs leading-snug">
                {group.description}
              </p>
              <ul className="mt-4 flex flex-wrap gap-3">
                {group.items.map((name) => (
                  <li key={name}>
                    <span
                      className="flex size-12 items-center justify-center rounded-lg border border-[#e5e5e5] bg-[#fafafa] transition-colors hover:border-[#d4d4d4] hover:bg-white"
                      title={name}
                    >
                      <BrandIcon name={name} size={24} slug={iconMap[name]} />
                    </span>
                  </li>
                ))}
              </ul>
            </motion.div>
          ))}
        </div>

        <motion.div
          animate={isInView ? { opacity: 1 } : {}}
          className="mt-14 flex justify-center md:mt-16"
          initial={{ opacity: 0 }}
          transition={{ duration: 0.4, delay: 0.35 }}
        >
          <p className="max-w-md text-center font-mono text-[#737373] text-[12px] leading-relaxed">
            Any runtime or CI that can run a binary and read env.
            <span className="mt-1 block text-[#a3a3a3]">
              No plugins required.
            </span>
          </p>
        </motion.div>
      </div>
    </section>
  );
}
