import { stripe } from '.';

export function handleWebhook(payload: string, signature: string) {
  const event = stripe.webhooks.constructEvent(
    payload,
    signature,
    // biome-ignore lint/style/noNonNullAssertion: Template file - env var will be set by user
    process.env.STRIPE_WEBHOOK_SECRET!
  );

  switch (event.type) {
    case 'invoice.payment_succeeded':
      // Handle successful payment
      break;
    case 'customer.subscription.deleted':
      // Handle subscription cancellation
      break;
    default:
      // Handle other events
      break;
  }

  return event;
}
