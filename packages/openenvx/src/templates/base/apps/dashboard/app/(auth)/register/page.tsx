import Link from 'next/link';
import { RegisterForm } from './register-form';

export default function RegisterPage() {
  return (
    <>
      <div className="flex flex-col space-y-2">
        <h1 className="font-semibold text-2xl tracking-tight">
          Create an account
        </h1>
        <p className="text-muted-foreground text-sm">
          Enter your details to get started
        </p>
      </div>

      <RegisterForm />

      <div className="space-y-4">
        <p className="text-center text-sm">
          Already have an account?{' '}
          <Link className="text-muted-foreground underline" href="/login">
            Sign in
          </Link>
        </p>
      </div>
    </>
  );
}
