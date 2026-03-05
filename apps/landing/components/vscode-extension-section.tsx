'use client';

import { ChevronDown, ChevronRight, FileCode, Folder } from 'lucide-react';
import { motion, useInView } from 'motion/react';
import { useRef } from 'react';

const leftItems = [
  {
    title: 'Script CodeLens',
    desc: 'Run npm scripts with OpenEnvX decrypted environment variables.',
  },
  {
    title: 'Secret Scanning',
    desc: 'Detect potential secret leaks before they reach git.',
  },
  {
    title: 'Go to Definition',
    desc: 'Jump from code to .env file definitions.',
  },
];

const rightItems = [
  {
    title: 'CodeLens Actions',
    desc: 'Decrypt individual secrets or all secrets at once with inline buttons.',
  },
  {
    title: 'Find References',
    desc: 'Monorepo-aware reference finding across packages and workspaces.',
  },
  {
    title: 'Rename Support',
    desc: 'Rename environment variables across your entire workspace.',
  },
];

export function VSCodeExtensionSection() {
  const sectionRef = useRef<HTMLElement>(null);
  const isInView = useInView(sectionRef, { amount: 0.08, once: true });

  return (
    <section
      aria-labelledby="vscode-extension-heading"
      className="relative py-24 md:py-32"
      ref={sectionRef}
    >
      <div className="mx-auto max-w-6xl px-6">
        <motion.div
          animate={isInView ? { opacity: 1, y: 0 } : {}}
          className="mb-12 text-left md:mb-16"
          initial={{ opacity: 0, y: 10 }}
          transition={{ duration: 0.4, ease: [0.25, 0.46, 0.45, 0.94] }}
        >
          <p className="mb-2 font-mono text-[#737373] text-xs uppercase tracking-widest">
            In your editor
          </p>
          <h2
            className="max-w-2xl font-semibold text-2xl text-black leading-tight tracking-tight md:text-3xl"
            id="vscode-extension-heading"
          >
            OpenEnvX for VS Code
          </h2>
          <p className="mt-3 max-w-xl text-[#525252] text-sm leading-relaxed">
            Edit .env with decrypted values, see status at a glance, and run
            OpenEnvX without leaving the editor.
          </p>
        </motion.div>

        <div className="grid items-start gap-8 xl:grid-cols-[200px_1fr_200px]">
          {/* Left column: feature list */}
          <motion.div
            animate={isInView ? { opacity: 1, x: 0 } : {}}
            className="order-1 grid gap-8 sm:grid-cols-2 md:grid-cols-3 xl:order-0 xl:grid-cols-1 xl:gap-0"
            initial={{ opacity: 0, x: -12 }}
            transition={{
              duration: 0.45,
              delay: 0.1,
              ease: [0.25, 0.46, 0.45, 0.94],
            }}
          >
            {leftItems.map((item, i) => (
              <div
                className={`flex flex-col gap-1 ${['xl:mt-0', 'xl:mt-10', 'xl:mt-24'][i] ?? 'xl:mt-10'}`}
                key={item.title}
              >
                <div className="flex flex-row items-center gap-2">
                  <p className="shrink-0 font-medium xl:text-sm">
                    {item.title}
                  </p>
                  <div className="hidden h-px flex-1 bg-black/10 xl:block" />
                </div>
                <p className="text-[#737373] xl:text-xs">{item.desc}</p>
              </div>
            ))}
          </motion.div>

          {/* Center: VS Code mock - sidebar | editor | panel */}
          <motion.div
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            className="hidden aspect-video overflow-hidden rounded-2xl border border-[#e5e5e5] bg-[#252526] md:grid md:grid-cols-[180px_1fr] md:divide-x md:divide-[#252525]"
            initial={{ opacity: 0, y: 16 }}
            transition={{
              duration: 0.5,
              delay: 0.15,
              ease: [0.25, 0.46, 0.45, 0.94],
            }}
          >
            {/* File tree */}
            <div className="flex flex-col gap-px py-4 pr-1.5 pl-1 font-mono text-[#8c8c8c] text-xs">
              <div className="flex items-center gap-1.5 rounded-md py-0.5 pl-1">
                <span className="flex w-4 shrink-0 justify-center">
                  <Folder className="size-3 text-[#8c8c8c]" strokeWidth={2} />
                </span>
                <span className="font-medium text-[#cccccc]">my-app</span>
              </div>
              <div className="flex items-center gap-1.5 rounded-md py-0.5 pl-4">
                <span className="flex w-4 shrink-0 justify-center">
                  <FileCode className="size-3 text-[#4ec9b0]" />
                </span>
                <span className="text-[#4ec9b0]">.openenvx.yaml</span>
              </div>
              <div className="flex items-center gap-1.5 rounded-md py-0.5 pl-4">
                <span className="flex w-4 shrink-0 justify-center">
                  <FileCode className="size-3 text-[#ce9178]" />
                </span>
                <span className="text-[#ce9178]">audit.logl</span>
              </div>
              <div className="flex items-center gap-1.5 rounded-md bg-[#2a2d2e] py-0.5 pl-4">
                <span className="flex w-4 shrink-0 justify-center">
                  <span className="size-3 rounded-full bg-[#16a34a]" />
                </span>
                <span className="font-medium text-[#d4d4d4]">.env</span>
              </div>
            </div>

            {/* Editor: .env content with References inline below line 2 */}
            <div className="grid grid-rows-[auto_1fr_auto] overflow-hidden bg-[#1e1e1e]">
              <div className="flex border-[#252525] border-b bg-[#2d2d2d] px-2">
                <div className="flex items-center gap-2 border-[#16a34a] border-b-2 py-2">
                  <span className="font-mono text-[#d4d4d4] text-[11px]">
                    .env
                  </span>
                  <span className="rounded bg-[#16a34a]/20 px-1.5 font-mono text-[#16a34a] text-[10px]">
                    encrypted
                  </span>
                </div>
              </div>
              <div className="overflow-auto p-4 font-mono text-[#d4d4d4] text-[12px] leading-relaxed">
                <div className="flex">
                  <span className="select-none pr-4 text-[#858585]"> 1 </span>
                  <span className="text-[#9cdcfe]">DATABASE_URL</span>
                  <span className="text-[#d4d4d4]">=</span>
                  <span className="text-[#ce9178]">
                    {' '}
                    postgres://localhost:5432/app
                  </span>
                </div>
                <div className="flex">
                  <span className="select-none pr-4 text-[#858585]"> 2 </span>
                  <span className="text-[#9cdcfe]">STRIPE_API_KEY</span>
                  <span className="text-[#d4d4d4]">=</span>
                  <span className="text-[#ce9178]"> sk_live_••••••••</span>
                </div>
                {/* References inline below line 2 */}
                <div className="mt-1.5 flex min-h-0 flex-col rounded border border-[#333] bg-[#252526] font-mono text-[10px]">
                  <div className="flex shrink-0 items-center truncate border-[#333] border-b px-1.5 py-0.5">
                    <span className="truncate text-[#999]">
                      .env &gt; STRIPE_API_KEY - References (2)
                    </span>
                  </div>
                  <div className="flex flex-col p-0.5">
                    <div className="flex flex-col gap-px">
                      <div className="flex items-center gap-0.5 py-px pr-0.5">
                        <ChevronDown className="size-2.5 shrink-0 text-[#666]" />
                        <FileCode className="size-2.5 shrink-0 text-[#9cdcfe]" />
                        <span className="truncate text-[#ccc]">
                          stripe-client.ts
                        </span>
                        <span className="ml-0.5 shrink-0 text-[#666]">
                          src/lib
                        </span>
                      </div>
                      <div className="ml-3.5 truncate rounded bg-[#094771]/80 py-px pr-1 pl-0.5 text-[#9cdcfe]">
                        createStripeClient(process.env.
                        <span className="text-[#d4d4d4]">STRIPE_API_KEY</span>)
                      </div>
                    </div>
                    <div className="flex items-center gap-0.5 py-px pr-0.5">
                      <ChevronRight className="size-2.5 shrink-0 text-[#666]" />
                      <FileCode className="size-2.5 shrink-0 text-[#9cdcfe]" />
                      <span className="truncate text-[#ccc]">billing.ts</span>
                      <span className="ml-0.5 shrink-0 text-[#666]">
                        src/features
                      </span>
                    </div>
                  </div>
                </div>
                <div className="mt-2 flex">
                  <span className="select-none pr-4 text-[#858585]"> 3 </span>
                  <span className="text-[#9cdcfe]">NODE_ENV</span>
                  <span className="text-[#d4d4d4]">=</span>
                  <span className="text-[#4ec9b0]"> development</span>
                </div>
              </div>
              <div className="flex items-center justify-between border-[#252525] border-t bg-[#007acc] px-3 py-1 font-mono text-[11px] text-white">
                <span>Ln 2, Col 1</span>
                <span className="flex items-center gap-1">
                  <span className="size-2 rounded-full bg-[#16a34a]" />
                  openenvx
                </span>
              </div>
            </div>
          </motion.div>

          {/* Right column: feature list */}
          <motion.div
            animate={isInView ? { opacity: 1, x: 0 } : {}}
            className="grid gap-8 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-1 xl:gap-0 xl:text-right"
            initial={{ opacity: 0, x: 12 }}
            transition={{
              duration: 0.45,
              delay: 0.2,
              ease: [0.25, 0.46, 0.45, 0.94],
            }}
          >
            {rightItems.map((item, i) => (
              <div
                className={`flex flex-col gap-1 ${['xl:mt-0', 'xl:mt-10', 'xl:mt-24'][i] ?? 'xl:mt-10'}`}
                key={item.title}
              >
                <div className="flex flex-row items-center gap-2 xl:flex-row-reverse">
                  <p className="shrink-0 font-medium xl:text-sm">
                    {item.title}
                  </p>
                  <div className="hidden h-px flex-1 bg-black/10 xl:block" />
                </div>
                <p className="text-[#737373] xl:text-xs">{item.desc}</p>
              </div>
            ))}
          </motion.div>
        </div>

        {/* Fallback: show mock on small screens as single column */}
        <div className="mt-8 overflow-hidden rounded-2xl border border-[#e5e5e5] bg-[#252526] md:hidden">
          <div className="flex flex-col">
            <div className="border-[#252525] border-b bg-[#2d2d2d] px-3 py-2">
              <span className="font-mono text-[#d4d4d4] text-xs">.env</span>
              <span className="ml-2 rounded bg-[#16a34a]/20 px-1.5 font-mono text-[#16a34a] text-[10px]">
                decrypted
              </span>
            </div>
            <pre className="p-4 font-mono text-[#d4d4d4] text-[11px] leading-relaxed">
              <span className="text-[#9cdcfe]">DATABASE_URL</span>
              <span className="text-[#d4d4d4]">=</span>
              <span className="text-[#ce9178]"> postgres://...</span>
              {'\n'}
              <span className="text-[#9cdcfe]">API_KEY</span>
              <span className="text-[#d4d4d4]">=</span>
              <span className="text-[#ce9178]"> sk_live_••••••••</span>
            </pre>
            <div className="flex items-center gap-2 border-[#252525] border-t bg-[#007acc] px-3 py-2 text-[11px] text-white">
              <span className="size-2 rounded-full bg-[#16a34a]" />
              openenvx
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
