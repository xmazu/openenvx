'use client';

import { ContentSection } from '@/components/content-section';
import { Features } from '@/components/features';
import { Footer } from '@/components/footer';
import { Header } from '@/components/header';
import { HeroSection } from '@/components/hero-section';
import { ManifestoSection } from '@/components/manifesto-section';
import { SafeToCommitSection } from '@/components/safe-to-commit-section';
import { VSCodeExtensionSection } from '@/components/vscode-extension-section';
import { WorksWithSection } from '@/components/works-with-section';

export default function Home() {
  return (
    <div className="min-h-screen bg-white">
      <Header />

      <main>
        <HeroSection />
        <Features />
        <SafeToCommitSection />
        <ContentSection />
        <VSCodeExtensionSection />
        <ManifestoSection />
        <WorksWithSection />
      </main>

      <Footer />
    </div>
  );
}
