export function Header() {
  return (
    <header className="sticky top-0 z-50 border-[#e5e5e5]/50 border-b bg-white/80 backdrop-blur-md">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
        <div className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-md bg-black">
            <span className="font-bold font-mono text-white text-xs">OX</span>
          </div>
          <span className="font-mono font-semibold text-sm">openenvx</span>
        </div>
        <nav className="flex items-center gap-1">
          <a
            className="rounded-md px-3 py-1.5 text-[#525252] text-sm transition-colors hover:bg-[#f5f5f5] hover:text-black"
            href="https://github.com/xmazu/openenvx"
            rel="noopener noreferrer"
            target="_blank"
          >
            GitHub
          </a>
        </nav>
      </div>
    </header>
  );
}
