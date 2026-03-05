import type { Metadata } from 'next';
import { Inter, JetBrains_Mono } from 'next/font/google';
import './globals.css';

const inter = Inter({
  variable: '--font-inter',
  subsets: ['latin'],
});

const jetbrainsMono = JetBrains_Mono({
  variable: '--font-jetbrains-mono',
  subsets: ['latin'],
});

export const metadata: Metadata = {
  title: 'OpenEnvX - .env encrypted. Zero config. Local-first.',
  description:
    'Zero-configuration secret management with envelope encryption. Your keys, your data, your control.',
  keywords: [
    'secrets',
    'env',
    'encryption',
    'cli',
    'security',
    'developer tools',
  ],
  authors: [{ name: 'OpenEnvX' }],
  openGraph: {
    title: 'OpenEnvX - .env encrypted. Zero config. Local-first.',
    description:
      'Zero-configuration secret management with envelope encryption. Your keys, your data, your control.',
    type: 'website',
    url: 'https://openenvx.dev',
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${inter.variable} ${jetbrainsMono.variable} bg-white text-black antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
