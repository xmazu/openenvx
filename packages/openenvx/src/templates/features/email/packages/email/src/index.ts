import { Resend } from 'resend';

export const resend = new Resend(process.env.RESEND_API_KEY);

// biome-ignore lint/performance/noBarrelFile: Template file - barrel exports are intentional for generated projects
export * from './templates';
