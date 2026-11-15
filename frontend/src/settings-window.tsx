import { Window } from '@wailsio/runtime';
import { BuiltInDefaults, Product } from '#bindings/desktop/service.ts';
import { mountWindow } from './mount-window.tsx';
import { SettingsWindow } from './settings/settings-window-view.tsx';
import { goTrustedKeys } from './settings/trusted-keys-source.ts';
import { TrustedKeysStore } from './settings/trusted-keys-store.ts';

// Escape closes the window unless an open popup took the key first.
document.addEventListener('keydown', (event) => {
  if (event.key === 'Escape' && !event.defaultPrevented) {
    void Window.Close();
  }
});

void Promise.all([Product(), BuiltInDefaults(), TrustedKeysStore.open(goTrustedKeys)]).then(
  ([product, builtInDefaults, trustedKeys]) =>
    mountWindow((stores) => (
      <SettingsWindow
        {...stores}
        product={product}
        builtInDefaults={builtInDefaults}
        trustedKeys={trustedKeys}
      />
    )),
);
