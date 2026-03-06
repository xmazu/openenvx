import { env } from '@{{projectName}}/env';
import { Resend } from 'resend';

export const resend = new Resend(env.RESEND_API_KEY);

// biome-ignore lint/performance/noBarrelFile: Template file - barrel exports are intentional for generated projects
export * from './templates';
