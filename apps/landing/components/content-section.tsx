'use client';

import { Check, X } from 'lucide-react';
import { motion, useInView } from 'motion/react';
import { useRef } from 'react';
import { SlackMock } from '@/components/slack-mock';

export function ContentSection() {
  const sectionRef = useRef<HTMLElement>(null);
  const isInView = useInView(sectionRef, { amount: 0.2, once: true });

  return (
    <section
      aria-labelledby="content-section-heading"
      className="section-grain relative overflow-hidden py-24 md:py-32"
      ref={sectionRef}
    >
      <div aria-hidden className="absolute inset-0 bg-[#fafafa]" />
      <div
        aria-hidden
        className="absolute top-0 bottom-0 left-0 w-px bg-linear-to-b from-transparent via-[#e5e5e5] to-transparent"
      />

      <div className="relative mx-auto max-w-6xl px-6">
        {/* Lead block */}
        <div className="max-w-3xl">
          <motion.div
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            initial={{ opacity: 0, y: 16 }}
            transition={{ duration: 0.5, ease: [0.25, 0.46, 0.45, 0.94] }}
          >
            <h2
              className="font-medium text-3xl text-black leading-[1.22] tracking-tight md:text-4xl lg:text-[2.75rem]"
              id="content-section-heading"
            >
              Secrets belong in encryption,
              <br />
              <span className="text-[#525252]">not in chat.</span>
            </h2>
            <p className="mt-6 max-w-xl text-[#737373] text-base leading-[1.6]">
              OpenEnvX keeps .env as ciphertext. Keys and ciphertext stay where
              you put them - so nothing leaks into Slack or commit history.
            </p>
          </motion.div>
        </div>

        {/* Card + Slack block */}
        <motion.div
          animate={isInView ? { opacity: 1 } : {}}
          className="mt-16 md:mt-20"
          initial={{ opacity: 0 }}
          transition={{ duration: 0.4, delay: 0.2 }}
        >
          <div className="grid gap-px overflow-hidden rounded-xl bg-[#e5e5e5] md:grid-cols-2">
            <div className="flex flex-col justify-center rounded-t-xl bg-white p-8 md:rounded-l-xl md:rounded-tr-none md:p-12">
              <div className="mb-5 flex h-11 w-11 items-center justify-center rounded-full bg-[#611f69]/10">
                <X aria-hidden className="size-5 text-[#611f69]" />
              </div>
              <h3 className="font-semibold text-black text-xl tracking-tight md:text-2xl">
                Stop sharing secrets in Slack
              </h3>
              <p className="mt-3 max-w-md text-[#525252] text-[15px] leading-[1.55]">
                Every day, developers share API keys, database credentials, and
                secrets over Slack, email, and DMs. Stop the leak before it
                happens.
              </p>
              <ul className="mt-6 flex flex-col gap-3">
                <li className="flex items-start gap-3 text-[#525252] text-sm leading-normal">
                  <Check
                    aria-hidden
                    className="mt-0.5 size-4 shrink-0 text-[#16a34a]"
                  />
                  <span>Encrypt once, decrypt only where you run</span>
                </li>
                <li className="flex items-start gap-3 text-[#525252] text-sm leading-normal">
                  <Check
                    aria-hidden
                    className="mt-0.5 size-4 shrink-0 text-[#16a34a]"
                  />
                  <span>Full audit trail for every decrypt</span>
                </li>
                <li className="flex items-start gap-3 text-[#525252] text-sm leading-normal">
                  <Check
                    aria-hidden
                    className="mt-0.5 size-4 shrink-0 text-[#16a34a]"
                  />
                  <span>No more pasting secrets into Slack</span>
                </li>
              </ul>
            </div>
            <div className="min-h-[380px] overflow-hidden rounded-b-xl border-[#e5e5e5] border-l bg-white md:min-h-[400px] md:rounded-r-xl md:rounded-bl-none">
              <SlackMock />
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
