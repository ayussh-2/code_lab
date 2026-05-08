export function SiteFooter() {
  return (
    <footer className="mt-auto bg-white shadow-[0px_0px_0px_1px_rgba(0,0,0,0.08)]">
      <div className="mx-auto flex w-full max-w-[1200px] flex-col items-center justify-between gap-4 px-6 py-8 md:flex-row">
        <div className="text-base font-medium text-[#171717]">CODE_LAB</div>
        <div className="flex flex-wrap justify-center gap-4 text-xs uppercase tracking-wider text-[#666666]">
          <span>Help Center</span>
          <span>Terms</span>
          <span>Privacy</span>
          <span>Jobs</span>
          <span>API</span>
        </div>
        <div className="text-sm text-[#4d4d4d]">© 2024 CODE_LAB. Built for technical precision.</div>
      </div>
    </footer>
  );
}
