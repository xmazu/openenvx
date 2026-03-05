import { Container, Heading, Html, Text } from '@react-email/components';

interface PaymentConfirmationProps {
  amount: string;
  date: string;
}

export function PaymentConfirmationEmail({
  amount,
  date,
}: PaymentConfirmationProps) {
  return (
    <Html>
      <Container>
        <Heading>Payment Confirmed</Heading>
        <Text>Amount: {amount}</Text>
        <Text>Date: {date}</Text>
      </Container>
    </Html>
  );
}
