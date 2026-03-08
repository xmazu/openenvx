import Link from 'next/link';
import { ForgotPasswordForm } from './forgot-password-form';

export default function ForgotPasswordPage() {
  return (
    <>
      <div className="flex flex-col space-y-2">
        <h1 className="font-semibold text-2xl tracking-tight">
          Reset your password
        </h1>
        <p className="text-muted-foreground text-sm">
          Enter your email address and we&apos;ll send you a link to reset your
          password
        </p>
      </div>

      <ForgotPasswordForm />

      <div className="space-y-4">
        <Link
          className="block text-center text-muted-foreground text-sm underline"
          href="/login"
        >
          Back to login
        </Link>
      </div>
    </>
  );
}
