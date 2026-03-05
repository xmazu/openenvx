'use client';

import { FileCode, Folder, Key, Lock, Terminal } from 'lucide-react';
import { motion, useInView } from 'motion/react';
import { useRef } from 'react';

const pillars = [
  {
    icon: Terminal,
    title: 'Zero config',
    line: 'init, set, run.',
  },
  {
    icon: Lock,
    title: 'AES-256-GCM',
    line: 'One DEK per secret. Your key wraps them all.',
  },
  {
    icon: Key,
    title: 'You hold the key',
    line: 'Master key stays local. Ciphertext can go anywhere.',
  },
];

export function Features() {
  const sectionRef = useRef<HTMLElement>(null);
  const isInView = useInView(sectionRef, { amount: 0.08, once: true });

  return (
    <section
      aria-labelledby="features-heading"
      className="section-grain relative flex flex-col overflow-hidden py-28 md:py-36"
      ref={sectionRef}
    >
      <div aria-hidden className="absolute inset-0 bg-[#fafafa]" />
      {/* Subtle top-edge fade for depth */}
      <div
        aria-hidden
        className="absolute inset-x-0 top-0 h-32 bg-linear-to-b from-white/60 to-transparent"
      />

      <div className="relative mx-auto w-full max-w-6xl px-6">
        {/* Lead: one line that sticks */}
        <motion.div
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          className="flex flex-col items-start gap-5 text-left"
          initial={{ opacity: 0, y: 12 }}
          transition={{ duration: 0.5, ease: [0.25, 0.46, 0.45, 0.94] }}
        >
          <p className="font-mono text-[#737373] text-xs uppercase tracking-widest">
            Why openenvx
          </p>
          <h2
            className="max-w-3xl font-semibold text-3xl text-black leading-[1.15] tracking-tight md:text-4xl lg:text-5xl"
            id="features-heading"
          >
            Control, transparency, trust.
          </h2>
          <p className="max-w-2xl text-[#525252] text-base leading-relaxed">
            One key per secret - you hold the master. Your repo, your key - no
            middleman, no API to call.
          </p>
        </motion.div>

        {/* Two hero cards */}
        <div className="mt-20 grid grid-cols-1 gap-8 md:grid-cols-2 md:gap-10">
          {/* Card 1: Envelope - accent on hover */}
          <motion.div
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            className="group relative flex flex-col gap-8 overflow-hidden rounded-3xl border border-[#e5e5e5] bg-white p-6 shadow-[0_1px_2px_rgba(0,0,0,0.04)] transition-[border-color,box-shadow] duration-300 hover:border-[#16a34a]/25 hover:shadow-[0_8px_30px_rgba(0,0,0,0.06)] md:p-8"
            initial={{ opacity: 0, y: 20 }}
            transition={{
              duration: 0.5,
              delay: 0.1,
              ease: [0.25, 0.46, 0.45, 0.94],
            }}
          >
            <div
              aria-hidden
              className="absolute top-0 bottom-0 left-0 w-1 rounded-l-3xl bg-[#16a34a] opacity-0 transition-opacity duration-300 group-hover:opacity-100"
            />
            <div className="relative">
              {/* Diagram: repo (ciphertext) | wraps | key (local) */}
              <div className="relative flex min-h-[200px] flex-col gap-4 md:min-h-[180px] md:flex-row md:items-stretch md:gap-0">
                {/* Left: .env in repo - file window */}
                <motion.div
                  animate={isInView ? { opacity: 1, x: 0 } : {}}
                  className="flex flex-1 flex-col overflow-hidden rounded-xl border border-[#e5e5e5] bg-[#fafafa] transition-all duration-300 group-hover:border-[#d4d4d4]"
                  initial={{ opacity: 0, x: -8 }}
                  transition={{ duration: 0.4, delay: 0.15 }}
                >
                  <div className="flex items-center justify-between border-[#e5e5e5] border-b bg-white/80 px-3 py-2">
                    <span className="font-mono text-[#737373] text-[11px]">
                      .env
                    </span>
                    <span className="rounded bg-[#e5e5e5] px-1.5 py-0.5 font-mono text-[#525252] text-[9px] uppercase tracking-wider">
                      ciphertext
                    </span>
                  </div>
                  <div className="flex flex-1 border-[#e5e5e5] border-t font-mono text-[10px] leading-relaxed">
                    <div className="shrink-0 select-none bg-white/60 py-2 pr-2 pl-3 text-right text-[#a3a3a3]">
                      1<br />2<br />3
                    </div>
                    <div className="min-w-0 bg-white/60 py-2 pr-3 text-[#737373]">
                      <div>DATABASE_URL=envx:…</div>
                      <div>API_KEY=envx:…</div>
                      <div>SECRET=envx:…</div>
                    </div>
                  </div>
                </motion.div>

                {/* Center: "wraps" connector */}
                <div className="flex shrink-0 items-center justify-center md:w-14">
                  <div className="flex flex-col items-center gap-1">
                    <div
                      aria-hidden
                      className="hidden h-8 w-px bg-linear-to-b from-transparent via-[#16a34a]/40 to-transparent md:block"
                    />
                    <Lock
                      className="size-4 shrink-0 text-[#16a34a]/70"
                      strokeWidth={2}
                    />
                    <span className="font-mono text-[#a3a3a3] text-[9px] uppercase tracking-widest">
                      wraps
                    </span>
                    <div
                      aria-hidden
                      className="hidden h-8 w-px bg-linear-to-b from-transparent via-[#16a34a]/40 to-transparent md:block"
                    />
                  </div>
                </div>

                {/* Right: age key - local only */}
                <motion.div
                  animate={isInView ? { opacity: 1, x: 0 } : {}}
                  className="flex flex-1 flex-col overflow-hidden rounded-xl border border-[#16a34a]/20 bg-[#f0fdf4]/50 transition-all duration-300 group-hover:border-[#16a34a]/35 group-hover:bg-[#f0fdf4]/70"
                  initial={{ opacity: 0, x: 8 }}
                  transition={{ duration: 0.4, delay: 0.2 }}
                >
                  <div className="flex items-center justify-between border-[#16a34a]/15 border-b bg-white/90 px-3 py-2">
                    <span className="font-mono text-[11px] text-black">
                      age key
                    </span>
                    <span className="flex items-center gap-1 rounded-full bg-[#16a34a]/15 px-2 py-0.5 font-mono text-[#15803d] text-[9px]">
                      <span className="size-1.5 rounded-full bg-[#16a34a]" />
                      local only
                    </span>
                  </div>
                  <div className="flex flex-1 flex-col justify-center gap-2 px-3 py-4">
                    <div className="flex items-center gap-2 text-[#525252] text-[10px]">
                      <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-[#16a34a]" />
                      Wraps each DEK
                    </div>
                    <div className="flex items-center gap-2 text-[#525252] text-[10px]">
                      <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-[#16a34a]" />
                      Stays on your machine
                    </div>
                  </div>
                  <div className="flex gap-2 border-[#16a34a]/10 border-t px-3 py-2">
                    <div className="flex-1 rounded-md bg-white/80 px-2 py-1.5 font-mono text-[9px]">
                      <span className="text-[#737373]">repo</span>
                      <span className="block truncate text-black">
                        .openenvx.yaml
                      </span>
                    </div>
                    <div className="flex-1 rounded-md bg-white/80 px-2 py-1.5 font-mono text-[9px]">
                      <span className="text-[#737373]">you keep</span>
                      <span className="block truncate text-black">
                        private key
                      </span>
                    </div>
                  </div>
                </motion.div>
              </div>
            </div>
            <div className="relative flex flex-col gap-2">
              <h3 className="font-semibold text-black text-xl">
                Envelope encryption
              </h3>
              <p className="text-[#525252] text-sm leading-relaxed">
                Each secret gets a unique data encryption key (DEK), then
                AES-256-GCM. Your age-derived master key wraps every DEK. Key
                names are associated data - no replay, no swap, no backdoor.
              </p>
            </div>
          </motion.div>

          {/* Card 2: Local-first - different hover */}
          <motion.div
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            className="group relative flex flex-col gap-8 overflow-hidden rounded-3xl border border-[#e5e5e5] bg-white p-6 shadow-[0_1px_2px_rgba(0,0,0,0.04)] transition-[border-color,box-shadow] duration-300 hover:border-[#d4d4d4] hover:shadow-[0_8px_30px_rgba(0,0,0,0.06)] md:p-8"
            initial={{ opacity: 0, y: 20 }}
            transition={{
              duration: 0.5,
              delay: 0.18,
              ease: [0.25, 0.46, 0.45, 0.94],
            }}
          >
            <div className="relative grid aspect-2/1 grid-cols-5 gap-3">
              <div className="col-span-3 rounded-2xl border border-[#e5e5e5] bg-[#fafafa]/90 p-3 shadow-md backdrop-blur transition-all duration-300 group-hover:-translate-x-0.5 group-hover:-translate-y-0.5 sm:p-4">
                <div className="flex items-center justify-between text-black text-xs">
                  <span>Your repo</span>
                  <span className="flex items-center gap-1 rounded-full bg-[#e8f5e9] px-2 py-0.5 font-mono text-[#2e7d32] text-[10px]">
                    <span className="h-1.5 w-1.5 rounded-full bg-[#2e7d32]" />
                    local
                  </span>
                </div>
                <div className="mt-3 rounded-lg border border-[#e5e5e5] bg-white px-3 py-2.5 font-mono text-[#525252] text-[11px]">
                  <div className="flex items-center gap-1.5 py-0.5">
                    <Folder
                      className="size-3 shrink-0 text-[#737373]"
                      strokeWidth={2}
                    />
                    <span>.cursor</span>
                  </div>
                  <div className="flex items-center gap-1.5 py-0.5">
                    <Folder
                      className="size-3 shrink-0 text-[#737373]"
                      strokeWidth={2}
                    />
                    <span>node_modules</span>
                  </div>
                  <div className="flex items-center gap-1.5 py-0.5">
                    <Folder
                      className="size-3 shrink-0 text-[#737373]"
                      strokeWidth={2}
                    />
                    <span>src</span>
                  </div>
                  <div className="flex items-center gap-1.5 py-0.5">
                    <FileCode className="size-3 shrink-0 text-[#16a34a]" />
                    <span>.env</span>
                    <span className="ml-1 text-[#16a34a] text-[10px]">
                      encrypted
                    </span>
                  </div>
                  <div className="flex items-center gap-1.5 py-0.5">
                    <FileCode className="size-3 shrink-0 text-[#737373]" />
                    <span>.openenvx.yaml</span>
                  </div>
                </div>
              </div>
              <div className="col-span-2 rounded-2xl border border-[#e5e5e5]/80 bg-[#fafafa]/50 p-3 text-[#737373] text-[11px] transition-all duration-300 group-hover:translate-x-0.5 group-hover:translate-y-0.5">
                <p className="mb-2 text-[#525252] text-xs">You own the flow</p>
                <ul className="space-y-1.5">
                  <li className="flex items-center gap-1.5">
                    <span className="h-1 w-1 rounded-full bg-[#16a34a]" />
                    Key stays local
                  </li>
                  <li className="flex items-center gap-1.5">
                    <span className="h-1 w-1 rounded-full bg-[#16a34a]" />
                    Same binary, dev & CI
                  </li>
                  <li className="flex items-center gap-1.5">
                    <span className="h-1 w-1 rounded-full bg-[#16a34a]" />
                    Commit .env safely
                  </li>
                  <li className="flex items-center gap-1.5">
                    <span className="h-1 w-1 rounded-full bg-[#16a34a]" />
                    No subscription
                  </li>
                  <li className="flex items-center gap-1.5">
                    <span className="h-1 w-1 rounded-full bg-[#16a34a]" />
                    No lock-in
                  </li>
                </ul>
              </div>
            </div>
            <div className="relative flex flex-col gap-2">
              <h3 className="font-semibold text-black text-xl">Local-first</h3>
              <p className="text-[#525252] text-sm leading-relaxed">
                Keys and ciphertext live where you put them. Same binary on your
                laptop, in CI, or on a server - you decide where to decrypt.
              </p>
            </div>
          </motion.div>
        </div>

        {/* Three pillars: list feel, not three boxes */}
        <div className="mt-16 grid grid-cols-1 gap-6 sm:grid-cols-3">
          {pillars.map((item, i) => (
            <motion.div
              animate={isInView ? { opacity: 1, y: 0 } : {}}
              className="group relative flex flex-col gap-4 rounded-2xl border border-[#e5e5e5] bg-white p-5 transition-colors hover:border-[#d4d4d4] hover:bg-[#fafafa]/80 md:p-6"
              initial={{ opacity: 0, y: 12 }}
              key={item.title}
              transition={{
                duration: 0.4,
                delay: 0.28 + i * 0.06,
                ease: [0.25, 0.46, 0.45, 0.94],
              }}
            >
              <div className="flex h-9 w-9 items-center justify-center rounded-xl border border-[#e5e5e5] bg-[#fafafa] text-[#525252] transition-colors group-hover:border-[#16a34a]/30 group-hover:bg-[#f0fdf4] group-hover:text-[#16a34a]">
                <item.icon className="size-4" strokeWidth={1.75} />
              </div>
              <div className="flex flex-col gap-1">
                <h3 className="font-semibold text-base text-black">
                  {item.title}
                </h3>
                <p className="text-[#737373] text-sm leading-relaxed">
                  {item.line}
                </p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
