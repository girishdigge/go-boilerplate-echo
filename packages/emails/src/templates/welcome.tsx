import {
  Body,
  Button,
  Container,
  Head,
  Heading,
  Hr,
  Html,
  Link,
  Preview,
  Section,
  Tailwind,
  Text,
} from 'react-email';

interface WelcomeEmailProps {
  userFirstName: string;
}

export const WelcomeEmail = ({
  userFirstName = '{{.UserFirstName}}',
}: WelcomeEmailProps) => {
  return (
    <Html lang='en'>
      <Preview>Welcome to Boilerplate!</Preview>

      <Tailwind>
        <Head />

        <Body className='bg-gray-100 font-sans'>
          <Container className='mx-auto my-10 max-w-[600px] rounded-lg bg-white p-8'>
            <Heading className='mt-4 text-2xl font-bold text-gray-800'>
              Welcome to Boilerplate!
            </Heading>

            <Section>
              <Text className='text-base text-gray-700'>
                Hi {userFirstName},
              </Text>

              <Text className='text-base text-gray-700'>
                Thank you for joining us. We're excited to have you on board!
              </Text>
            </Section>

            <Section className='my-8 text-center'>
              <Button
                className='rounded-md bg-orange-600 px-6 py-3 font-medium text-white'
                href='/dashboard'
              >
                Get Started
              </Button>
            </Section>

            <Hr className='my-6 border-gray-200' />

            <Section>
              <Text className='text-sm text-gray-600'>
                If you have any questions, feel free to{' '}
                <Link href='/support' className='text-orange-600 underline'>
                  contact our support team
                </Link>
                .
              </Text>
            </Section>

            <Section className='mt-8 text-center'>
              <Text className='text-xs text-gray-500'>
                © {new Date().getFullYear()} Boilerplate. All rights reserved.
              </Text>
            </Section>
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
};

WelcomeEmail.PreviewProps = {
  userFirstName: 'Girish',
};

export default WelcomeEmail;
