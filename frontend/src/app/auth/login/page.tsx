'use client';

import Layout from '@/components/Layout';
import LoginForm from '@/components/auth/LoginForm';

export default function LoginPage() {
  return (
    <Layout>
      <div className="flex min-h-screen flex-col items-center justify-center p-24">
        <LoginForm />
      </div>
    </Layout>
  );
}
