"use client";

import { useState } from "react";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const FORMATS = ["jpeg", "png", "gif", "webp"];

type Result = { url: string; name: string; before: number; after: number };

const kb = (n: number) => `${(n / 1024).toFixed(1)} KB`;
const maxFileSize = 20 * 1024 * 1024; // 20 MB

export default function Home() {
  const [format, setFormat] = useState("jpeg");
  const [quality, setQuality] = useState(80);
  const [webpLossy, setWebpLossy] = useState(false);
  const [result, setResult] = useState<Result | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const lossy = format === "jpeg" || (format === "webp" && webpLossy);

  async function convert(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const body = new FormData(e.currentTarget);
    body.append("lossy", String(lossy));
    const file = body.get("file") as File;
    setBusy(true);
    setError("");
    try {
      if (file.size > maxFileSize) {
        throw new Error(
          `File size exceeds the maximum limit of ${kb(maxFileSize)}`,
        );
      }

      const res = await fetch(`${API}/convert`, { method: "POST", body });
      if (!res.ok) throw new Error(await res.text());
      const blob = await res.blob();
      if (result) URL.revokeObjectURL(result.url);
      setResult({
        url: URL.createObjectURL(blob),
        name: `${file.name.replace(/\.[^.]+$/, "")}.${format === "jpeg" ? "jpg" : format}`,
        before: file.size,
        after: blob.size,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="mx-auto w-full max-w-xl p-6 sm:p-10">
      <h1 className="mb-6 text-2xl font-semibold">Image Converter</h1>

      <form onSubmit={convert} className="flex flex-col gap-5">
        <label className="flex flex-col gap-2">
          <span className="text-sm font-medium">Image</span>
          <input
            name="file"
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            required
            className="rounded border border-foreground/20 p-2 file:mr-3 file:rounded file:border-0 file:bg-foreground/10 file:px-3 file:py-1"
          />
        </label>

        <label className="flex flex-col gap-2">
          <span className="text-sm font-medium">Convert to</span>
          <select
            name="format"
            value={format}
            onChange={(e) => setFormat(e.target.value)}
            className="rounded border border-foreground/20 bg-background p-2"
          >
            {FORMATS.map((f) => (
              <option key={f} value={f}>
                {f.toUpperCase()}
              </option>
            ))}
          </select>
        </label>

        {format === "webp" && (
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={webpLossy}
              onChange={(e) => setWebpLossy(e.target.checked)}
            />
            <span className="text-sm font-medium">Lossy</span>
          </label>
        )}

        <label className="flex flex-col gap-2">
          <span className="text-sm font-medium">
            Quality: {lossy ? quality : "n/a (lossless)"}
          </span>
          <input
            name="quality"
            type="range"
            min={1}
            max={100}
            value={quality}
            disabled={!lossy}
            onChange={(e) => setQuality(Number(e.target.value))}
            className="disabled:opacity-40"
          />
        </label>

        <button
          disabled={busy}
          className="rounded bg-foreground px-4 py-2 font-medium text-background disabled:opacity-50"
        >
          {busy ? "Converting…" : "Convert"}
        </button>
      </form>

      {error && (
        <p role="alert" className="mt-6 text-red-600">
          {error}
        </p>
      )}

      {result && (
        <section className="mt-8 flex flex-col gap-3">
          <p className="text-sm">
            {kb(result.before)} → {kb(result.after)} (
            {Math.round(
              (result.after / result.before - 1) * 100,
            ).toLocaleString(undefined, { signDisplay: "exceptZero" })}
            %)
          </p>
          {/* eslint-disable-next-line @next/next/no-img-element -- blob URL, next/image can't optimize it */}
          <img
            src={result.url}
            alt="Converted result"
            className="max-h-96 rounded border border-foreground/20 object-contain"
          />
          <a
            href={result.url}
            download={result.name}
            className="self-start rounded border border-foreground/20 px-4 py-2 font-medium"
          >
            Download {result.name}
          </a>
        </section>
      )}
    </main>
  );
}
