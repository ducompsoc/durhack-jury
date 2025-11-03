"use client"

import { useSearchParams } from 'react-router-dom';

import { isValidStatusCode, statusMessages } from '../lib/status-codes';
import { ServerErrorPage } from '../components/ServerError';

export default function ErrorPage() {
  const [searchParams, setSearchParams] = useSearchParams()

  const statusCode = Number(searchParams.get("status_code"))
  const retryHref = searchParams.get("retry_href")

  if (isValidStatusCode(statusCode)) {
    return <ServerErrorPage statusCode={statusCode.toString()} message={statusMessages[statusCode]} retryHref={retryHref} />
  }

  if (!Number.isNaN(statusCode)) {
    return <ServerErrorPage statusCode={statusCode.toString()} message="Something strange..." retryHref={retryHref} />
  }

  return <ServerErrorPage message="Something strange..." retryHref={retryHref} />
}
