export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen">
      {/* Left side - Brand/Illustration */}
      <div className="relative hidden overflow-hidden bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 lg:flex lg:w-1/2">
        <div className="absolute inset-0 bg-[url('https://images.unsplash.com/photo-1557683316-973673baf926?w=1200')] bg-center bg-cover opacity-20" />
        <div className="absolute inset-0 bg-gradient-to-br from-slate-900/90 via-slate-800/80 to-transparent" />

        <div className="relative z-10 flex flex-col justify-between p-12 text-white">
          <div className="flex items-center gap-2">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-white font-bold text-slate-900 text-xl">
              E
            </div>
            <span className="font-semibold text-xl">{'{{projectName}}'}</span>
          </div>

          <div className="max-w-lg space-y-6">
            <h1 className="font-bold text-5xl leading-tight">
              Secure Environment
              <br /> Management
            </h1>
            <p className="text-slate-300 text-xl leading-relaxed">
              Manage your environment variables securely with end-to-end
              encryption.
            </p>
          </div>

          <p className="text-sm text-white/60">
            Secure by default, local-first, developer-friendly.
          </p>
        </div>
      </div>

      {/* Right side - Auth form */}
      <div className="flex flex-1 items-center justify-center bg-white p-8">
        <div className="w-full max-w-[450px] space-y-6">{children}</div>
      </div>
    </div>
  );
}
