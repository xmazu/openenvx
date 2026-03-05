import Stripe from 'stripe';

// biome-ignore lint/style/noNonNullAssertion: Template file - env var will be set by user
export const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
  apiVersion: '2024-04-10',
});

// biome-ignore lint/performance/noBarrelFile: Template file - barrel exports are intentional for generated projects
export * from './prices';
export * from './webhooks';
