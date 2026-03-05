import { Button, Container, Html, Text } from '@react-email/components';

interface WelcomeEmailProps {
  name: string;
  url: string;
}

export function WelcomeEmail({ name, url }: WelcomeEmailProps) {
  return (
    <Html>
      <Container>
        <Text>Welcome, {name}!</Text>
        <Text>Thanks for signing up.</Text>
        <Button href={url}>Get Started</Button>
      </Container>
    </Html>
  );
}
