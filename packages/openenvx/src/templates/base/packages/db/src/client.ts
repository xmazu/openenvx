import { neon } from '@neondatabase/serverless';
import { drizzle } from 'drizzle-orm/neon-http';

// biome-ignore lint/style/noNonNullAssertion: Template file - env var will be set by user
const sql = neon(process.env.DATABASE_URL!);
export const db = drizzle(sql);
