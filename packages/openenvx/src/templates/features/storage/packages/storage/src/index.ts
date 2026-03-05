import { PutObjectCommand, S3Client } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';

const s3Client = new S3Client({
  region: process.env.AWS_REGION || 'us-east-1',
  credentials: {
    // biome-ignore lint/style/noNonNullAssertion: Template file - env var will be set by user
    accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
    // biome-ignore lint/style/noNonNullAssertion: Template file - env var will be set by user
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!,
  },
});

// biome-ignore lint/style/noNonNullAssertion: Template file - env var will be set by user
const BUCKET_NAME = process.env.S3_BUCKET_NAME!;

export function getUploadUrl(
  key: string,
  contentType: string
): Promise<string> {
  const command = new PutObjectCommand({
    Bucket: BUCKET_NAME,
    Key: key,
    ContentType: contentType,
  });

  return getSignedUrl(s3Client, command, { expiresIn: 300 });
}
