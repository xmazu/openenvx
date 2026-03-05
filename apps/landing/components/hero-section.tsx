'use client';

import {
  AnimatedSpan,
  Terminal,
  TypingAnimation,
} from '@/components/termina-v2';

export function HeroSection() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-24 md:py-32">
      <div className="grid items-center gap-12 md:grid-cols-2">
        <div className="space-y-8">
          <div className="space-y-4">
            <h1 className="font-medium text-4xl leading-tight tracking-tight md:text-5xl lg:text-6xl">
              <span className="text-black">.env encrypted.</span>
              <br />
              <span className="text-[#525252]">Zero config.</span>
              <br />
              <span className="text-[#a3a3a3]">Local-first.</span>
            </h1>
            <p className="max-w-md text-[#737373] text-lg">
              Envelope encryption, zero config. One CLI, one key - you hold it.
            </p>
            <p className="text-[#a3a3a3] text-sm">
              Open source. Commit encrypted .env, decrypt only where you run.
            </p>
          </div>
          <div className="flex flex-wrap gap-4">
            <a
              className="inline-flex items-center gap-2 rounded-md bg-black px-5 py-2.5 font-medium text-sm text-white transition-colors hover:bg-[#171717]"
              href="https://github.com/xmazu/openenvx#quick-start"
              rel="noopener noreferrer"
              target="_blank"
            >
              Get Started
            </a>
            <a
              className="inline-flex items-center gap-2 rounded-md border border-[#d4d4d4] px-5 py-2.5 font-medium text-[#525252] text-sm transition-colors hover:border-[#a3a3a3] hover:text-black"
              href="https://github.com/xmazu/openenvx"
              rel="noopener noreferrer"
              target="_blank"
            >
              View on GitHub
            </a>
          </div>
        </div>
        <div className="flex justify-center md:justify-end">
          <Terminal className="max-h-80" loop>
            <TypingAnimation>$ openenvx init</TypingAnimation>
            <AnimatedSpan>
              <span className="text-[#16a34a]">✓ Generated keypair</span>
            </AnimatedSpan>
            <AnimatedSpan>
              <span className="text-[#16a34a]">✓ Encrypted .env</span>
            </AnimatedSpan>
            <TypingAnimation>$ openenvx set DATABASE_URL</TypingAnimation>
            <AnimatedSpan>
              <span className="text-[#737373]">Enter value: ••••••••</span>
            </AnimatedSpan>
            <AnimatedSpan>
              <span className="text-[#16a34a]">✓ Secret stored</span>
            </AnimatedSpan>
            <TypingAnimation>$ openenvx run -- next dev</TypingAnimation>
            <AnimatedSpan>
              <span className="text-[#2563eb]">
                Running with decrypted environment...
              </span>
            </AnimatedSpan>
          </Terminal>
        </div>
      </div>
    </section>
  );
}
