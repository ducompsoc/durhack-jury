import type React from "react"

const styles: Record<string, React.CSSProperties> = {
  error: {
    textAlign: "center",
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    justifyContent: "center",
  },
  desc: {
    display: "flex",
    height: "96px",
    flexDirection: "row",
    alignItems: "center",
  },
  h1: {
    display: "inline-block",
    margin: "0 20px 0 0",
    paddingRight: 23,
    fontSize: 48,
    fontWeight: 500,
    verticalAlign: "top",
    borderRight: "1px solid rgba(0, 0, 0, .3)",
  },
  h2: {
    fontSize: 28,
    fontWeight: 400,
    lineHeight: "28px",
  },
  wrap: {
    display: "inline-block",
  },
  retry: {
    width: "5rem",
    paddingTop: "0.5rem",
    paddingBottom: "0.5rem",
    borderRadius: "10px",
    backgroundColor: "var(--primary)",
    color: "white",
  }
}

type RetryButtonProps = {
  retryHref?: string | URL | null | undefined
} & React.ComponentProps<"a">

function RetryButton({ retryHref, style, ...props }: RetryButtonProps) {
  if (retryHref == null) return null
  return <a href={retryHref.toString()} style={{ ...styles.retry, ...style }} {...props}>
    Retry
  </a>
}

export type HttpErrorProps = {
  statusCode?: string | null | undefined
  message: string
  retryHref?: string | URL | null | undefined
} & React.ComponentProps<"div">

export function ServerErrorPage({ style, ...props }: HttpErrorProps) {
  style ??= {}
  style.minHeight = "100vh"
  return <ServerError style={style} {...props} />
}

export function ServerError({ statusCode, message, retryHref, style, ...props }: HttpErrorProps) {
  return (
    <main style={{ ...styles.error, ...style }} {...props}>
      <div style={styles.desc}>
        {statusCode ? <h1 style={styles.h1}>{statusCode}</h1> : null}
        <div style={styles.wrap}>
          <h2 style={styles.h2}>{message}.</h2>
        </div>
      </div>
      <RetryButton retryHref={retryHref} />
    </main>
  )
}
