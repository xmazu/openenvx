export function SafeToCommitSection() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-24">
      <div className="grid items-center gap-12 md:grid-cols-2">
        <div>
          <h2 className="mb-4 font-medium text-2xl md:text-3xl">
            Safe to commit
          </h2>
          <p className="mb-6 max-w-md text-[#737373]">
            .env holds only ciphertext. Without the private key it's unreadable.
            Store the key in 1Password or a hardware key. Commit .env to the
            repo without worry.
          </p>
          <ul className="space-y-3">
            <li className="flex items-start gap-3">
              <span className="mt-0.5 text-[#22c55e] text-sm">✓</span>
              <span className="text-[#525252] text-sm">
                Public key stored in .openenvx.yaml (safe to commit)
              </span>
            </li>
            <li className="flex items-start gap-3">
              <span className="mt-0.5 text-[#22c55e] text-sm">✓</span>
              <span className="text-[#525252] text-sm">
                Private key stored locally
              </span>
            </li>
            <li className="flex items-start gap-3">
              <span className="mt-0.5 text-[#22c55e] text-sm">✓</span>
              <span className="text-[#525252] text-sm">
                Environment variable override supported
              </span>
            </li>
          </ul>
        </div>
        <div className="overflow-hidden rounded-lg border border-[#e5e5e5] bg-[#fafafa]">
          <div className="flex items-center gap-2 border-[#e5e5e5] border-b bg-white px-4 py-3">
            <span className="font-mono text-[#a3a3a3] text-xs">.env</span>
          </div>
          <div className="p-4 font-mono text-sm leading-relaxed">
            <div className="text-[#525252]">
              <span className="text-[#2563eb]">DATABASE_URL</span>=
              <span className="text-[#a3a3a3]">envx:a2V5...:Y2lwaGVy...</span>
            </div>
            <div className="text-[#525252]">
              <span className="text-[#2563eb]">API_KEY</span>=
              <span className="text-[#a3a3a3]">envx:bDNr...:ZGF0YQ...</span>
            </div>
            <div className="text-[#525252]">
              <span className="text-[#2563eb]">SECRET_TOKEN</span>=
              <span className="text-[#a3a3a3]">envx:Y2Vk...:ZW5jcnlw...</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
