/**
 * Copyright 2026 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * Hub settings page component
 *
 * Displays hub-scoped environment variables and secrets.
 */

import { LitElement, html, css } from 'lit';
import { customElement } from 'lit/decorators.js';

import '../shared/env-var-list.js';
import '../shared/secret-list.js';

@customElement('scion-page-settings')
export class ScionPageSettings extends LitElement {
  static override styles = css`
    :host {
      display: block;
    }

    .header {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-bottom: 2rem;
    }

    .header sl-icon {
      color: var(--scion-primary, #3b82f6);
      font-size: 1.5rem;
    }

    .header h1 {
      font-size: 1.5rem;
      font-weight: 700;
      color: var(--scion-text, #1e293b);
      margin: 0;
    }
  `;

  override render() {
    return html`
      <div class="header">
        <sl-icon name="gear"></sl-icon>
        <h1>Hub Settings</h1>
      </div>

      <scion-env-var-list
        scope="hub"
        apiBasePath="/api/v1"
        compact
      ></scion-env-var-list>

      <scion-secret-list
        scope="hub"
        apiBasePath="/api/v1"
        compact
      ></scion-secret-list>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'scion-page-settings': ScionPageSettings;
  }
}
