import Link from 'next/link';
import { LoginForm } from './login-form';

export default function LoginPage() {
  return (
    <>
      <div className="flex flex-col space-y-2">
        <h1 className="font-semibold text-2xl tracking-tight">Welcome back</h1>
        <p className="text-muted-foreground text-sm">
          Enter your credentials to access your account
        </p>
      </div>

      <LoginForm />

      <div className="space-y-4">
        <Link
          className="block text-center text-muted-foreground text-sm underline"
          href="/forgot-password"
        >
          Forgot password?
        </Link>
        <p className="text-center text-sm">
          Don&apos;t have an account?{' '}
          <Link className="text-muted-foreground underline" href="/register">
            Create one
          </Link>
        </p>
      </div>
    </>
  );
}
