import { Browser } from '@wailsio/runtime';
import { useId } from 'react';
import type { Product } from '#bindings/product/models.ts';
import { Button } from '@/components/ui/button';
import { m } from '@/paraglide/messages.js';
import { SettingRow } from './setting-row.tsx';

/** The app itself: what it is, which version runs, its license, where its source lives and how to support it. */
export function AboutSection({ product }: { product: Product }) {
  const licenseId = useId();
  return (
    <section aria-labelledby="about-title" className="w-full max-w-2xl px-8 pb-8">
      <h1 id="about-title" className="text-lg font-semibold">
        {m.settings_about()}
      </h1>
      <div className="mt-4 flex items-center gap-4">
        {product.icon && (
          <img src={`data:image/png;base64,${product.icon}`} alt="" className="size-16" />
        )}
        <div className="flex flex-col gap-0.5">
          <span className="text-base font-semibold">{product.name}</span>
          <span className="text-sm text-muted-foreground">
            {m.about_version({ version: product.version })}
          </span>
          <span className="text-sm text-muted-foreground">{product.copyright}</span>
        </div>
      </div>
      <div data-surface="group" className="mt-6 divide-y rounded-lg border bg-card">
        <SettingRow labelId={licenseId} label={m.about_license()}>
          <span className="text-sm text-muted-foreground">{product.license}</span>
        </SettingRow>
        <LinkRow label={m.about_source()} url={product.repository} />
        <LinkRow label={m.about_donate()} url={product.donate} />
      </div>
    </section>
  );
}

/** A page outside the app: its address under the label and a button opening it in the system's browser. */
function LinkRow({ label, url }: { label: string; url: string }) {
  const id = useId();
  return (
    <SettingRow labelId={id} label={label} hint={url.replace(/^https:\/\//, '')}>
      <Button
        variant="secondary"
        size="sm"
        aria-describedby={id}
        onClick={() => void Browser.OpenURL(url)}
      >
        {m.about_open()}
      </Button>
    </SettingRow>
  );
}
