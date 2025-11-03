"use client"

import { useSearchParams } from 'react-router-dom';

import { isValidStatusCode, statusMessages } from '../lib/status-codes';
import { ServerErrorPage } from '../components/ServerError';

export default function ErrorPage() {
  const [searchParams, setSearchParams] = useSearchParams()

  const statusCode = Number(searchParams.get("status_code"))

  if (isValidStatusCode(statusCode)) {
    return <ServerErrorPage statusCode={statusCode.toString()} message={statusMessages[statusCode]} />
  }

  if (!Number.isNaN(statusCode)) {
    return <ServerErrorPage statusCode={statusCode.toString()} message="Something strange..." />
  }

  return <ServerErrorPage message="Something strange..." />
}