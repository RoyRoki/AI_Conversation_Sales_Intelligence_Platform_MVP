import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "AI Conversation & Sales Intelligence Platform",
  description: "Admin dashboard for conversation management and sales intelligence",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="bg-gray-900 text-white">
        {children}
      </body>
    </html>
  );
}
