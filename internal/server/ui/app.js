(() => {
  const app = document.querySelector("[data-app]");
  const flashRoot = document.querySelector("[data-flash]");
  const view = document.body.dataset.view;
  const repoOwner = document.body.dataset.repoOwner;
  const repoName = document.body.dataset.repoName;
  const logoutButton = document.querySelector('[data-action="logout"]');

  const escapeHTML = (value) =>
    String(value ?? "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");

  const formatDate = (value) => {
    if (!value) {
      return "Not available";
    }
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? "Not available" : date.toLocaleString();
  };

  const showFlash = (message, kind = "error") => {
    if (!flashRoot) {
      return;
    }
    flashRoot.innerHTML = message
      ? `<div class="flash ${escapeHTML(kind)}">${escapeHTML(message)}</div>`
      : "";
  };

  const requestJSON = async (url, options = {}) => {
    const headers = new Headers(options.headers || {});
    headers.set("Accept", "application/json");
    if (options.body && !headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }

    const response = await fetch(url, {
      ...options,
      headers,
      credentials: "same-origin",
    });

    let payload = null;
    const text = await response.text();
    if (text) {
      try {
        payload = JSON.parse(text);
      } catch {
        payload = { error: text };
      }
    }

    if (!response.ok) {
      const error = new Error(payload?.error || `Request failed with status ${response.status}`);
      error.status = response.status;
      throw error;
    }

    return payload || {};
  };

  const hydrateLogout = () => {
    if (!logoutButton) {
      return;
    }
    if (view === "login" || view === "register") {
      logoutButton.closest(".app-nav")?.remove();
      return;
    }
    logoutButton.addEventListener("click", async () => {
      try {
        await requestJSON("/api/v1/auth/logout", { method: "POST" });
      } catch {
        // Ignore logout failures and still reset the browser state.
      }
      window.location.href = "/app/login";
    });
  };

  const renderLogin = () => {
    app.innerHTML = `
      <section class="hero-panel">
        <p class="meta-text">Browser Access</p>
        <h1>Sign in to Forge.</h1>
        <p class="muted">Use the same account that already works for the JSON API and Git transport.</p>
      </section>
      <section class="panel">
        <form class="form-stack" data-form="login">
          <label class="field">
            <span>Username</span>
            <input name="username" autocomplete="username" required>
          </label>
          <label class="field">
            <span>Password</span>
            <input name="password" type="password" autocomplete="current-password" required>
          </label>
          <div class="button-row">
            <button class="button" type="submit">Sign In</button>
            <a class="button secondary" href="/app/register">Create Account</a>
          </div>
        </form>
      </section>
    `;

    app.querySelector('[data-form="login"]').addEventListener("submit", async (event) => {
      event.preventDefault();
      showFlash("");
      const form = new FormData(event.currentTarget);
      try {
        await requestJSON("/api/v1/auth/login", {
          method: "POST",
          body: JSON.stringify({
            username: form.get("username"),
            password: form.get("password"),
          }),
        });
        window.location.href = "/app/repos";
      } catch (error) {
        showFlash(error.message);
      }
    });
  };

  const renderRegister = () => {
    app.innerHTML = `
      <section class="hero-panel">
        <p class="meta-text">Create Account</p>
        <h1>Start hosting code you control.</h1>
        <p class="muted">Forge begins with a real Git backend. The browser layer simply makes it easier to use.</p>
      </section>
      <section class="panel">
        <form class="form-stack" data-form="register">
          <label class="field">
            <span>Username</span>
            <input name="username" autocomplete="username" required>
          </label>
          <label class="field">
            <span>Password</span>
            <input name="password" type="password" autocomplete="new-password" minlength="12" required>
          </label>
          <div class="button-row">
            <button class="button" type="submit">Create Account</button>
            <a class="button secondary" href="/app/login">Back to Sign In</a>
          </div>
        </form>
      </section>
    `;

    app.querySelector('[data-form="register"]').addEventListener("submit", async (event) => {
      event.preventDefault();
      showFlash("");
      const form = new FormData(event.currentTarget);
      try {
        await requestJSON("/api/v1/auth/register", {
          method: "POST",
          body: JSON.stringify({
            username: form.get("username"),
            password: form.get("password"),
          }),
        });
        window.location.href = "/app/repos";
      } catch (error) {
        showFlash(error.message);
      }
    });
  };

  const renderRepositories = async () => {
    const [{ user }, { repositories }, { organizations }] = await Promise.all([
      requestJSON("/api/v1/me"),
      requestJSON("/api/v1/repos"),
      requestJSON("/api/v1/orgs"),
    ]);

    const organizationOptions = organizations.length
      ? organizations
          .map(
            (organization) =>
              `<option value="${escapeHTML(organization.organization_slug)}">${escapeHTML(
                organization.organization_slug
              )} (${escapeHTML(organization.role)})</option>`
          )
          .join("")
      : "";

    const repoCards = repositories.length
      ? repositories
          .map(
            (repository) => `
              <article class="repo-card">
                <div class="button-row">
                  <span class="pill">${escapeHTML(repository.owner_type)}</span>
                  <span class="pill">${escapeHTML(repository.visibility)}</span>
                </div>
                <h3>${escapeHTML(repository.owner)}/${escapeHTML(repository.name)}</h3>
                <p class="muted">${escapeHTML(repository.description || "No description yet.")}</p>
                <p class="meta-text">Default branch: ${escapeHTML(repository.default_branch || "main")}</p>
                <a class="button secondary" href="/app/repos/${encodeURIComponent(repository.owner)}/${encodeURIComponent(
                  repository.name
                )}">Open Repository</a>
              </article>
            `
          )
          .join("")
      : `<div class="empty-state">No repositories are visible to this account yet.</div>`;

    app.innerHTML = `
      <section class="hero-panel">
        <p class="meta-text">Workbench</p>
        <h1>${escapeHTML(user.username)} can now manage repositories, organizations, and keys from the browser.</h1>
        <p class="muted">This UI stays intentionally close to the backend. It uses the same auth cookies and JSON routes as the API clients.</p>
      </section>
      <section class="stats-grid">
        <article class="stat-card">
          <h3>${repositories.length}</h3>
          <p class="muted">Accessible repositories</p>
        </article>
        <article class="stat-card">
          <h3>${organizations.length}</h3>
          <p class="muted">Organizations you belong to</p>
        </article>
        <article class="stat-card">
          <h3>${escapeHTML(user.role)}</h3>
          <p class="muted">Account role</p>
        </article>
      </section>
      <section class="split-grid">
        <article class="panel">
          <h2>Create Personal Repository</h2>
          <form class="form-stack" data-form="create-user-repo">
            <label class="field"><span>Name</span><input name="name" required></label>
            <label class="field"><span>Description</span><textarea name="description"></textarea></label>
            <label class="field">
              <span>Visibility</span>
              <select name="visibility">
                <option value="private">Private</option>
                <option value="public">Public</option>
              </select>
            </label>
            <label class="field"><span>Default branch</span><input name="default_branch" value="main" required></label>
            <button class="button" type="submit">Create Repository</button>
          </form>
        </article>
        <article class="panel">
          <h2>Create Organization</h2>
          <form class="form-stack" data-form="create-org">
            <label class="field"><span>Slug</span><input name="slug" required></label>
            <label class="field"><span>Display name</span><input name="display_name"></label>
            <label class="field"><span>Description</span><textarea name="description"></textarea></label>
            <button class="button" type="submit">Create Organization</button>
          </form>
        </article>
      </section>
      <section class="panel">
        <h2>Create Organization Repository</h2>
        ${
          organizations.length
            ? `<form class="form-stack" data-form="create-org-repo">
                <label class="field">
                  <span>Organization</span>
                  <select name="owner">${organizationOptions}</select>
                </label>
                <label class="field"><span>Name</span><input name="name" required></label>
                <label class="field"><span>Description</span><textarea name="description"></textarea></label>
                <label class="field">
                  <span>Visibility</span>
                  <select name="visibility">
                    <option value="private">Private</option>
                    <option value="public">Public</option>
                  </select>
                </label>
                <label class="field"><span>Default branch</span><input name="default_branch" value="main" required></label>
                <button class="button" type="submit">Create Organization Repository</button>
              </form>`
            : `<div class="empty-state">Create or join an organization first.</div>`
        }
      </section>
      <section class="panel">
        <h2>Repositories</h2>
        <div class="repo-grid">${repoCards}</div>
      </section>
    `;

    app.querySelector('[data-form="create-user-repo"]').addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await requestJSON("/api/v1/repos", {
          method: "POST",
          body: JSON.stringify({
            name: form.get("name"),
            description: form.get("description"),
            visibility: form.get("visibility"),
            default_branch: form.get("default_branch"),
          }),
        });
        window.location.reload();
      } catch (error) {
        showFlash(error.message);
      }
    });

    app.querySelector('[data-form="create-org"]').addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await requestJSON("/api/v1/orgs", {
          method: "POST",
          body: JSON.stringify({
            slug: form.get("slug"),
            display_name: form.get("display_name"),
            description: form.get("description"),
          }),
        });
        window.location.reload();
      } catch (error) {
        showFlash(error.message);
      }
    });

    const createOrgRepoForm = app.querySelector('[data-form="create-org-repo"]');
    if (createOrgRepoForm) {
      createOrgRepoForm.addEventListener("submit", async (event) => {
        event.preventDefault();
        const form = new FormData(event.currentTarget);
        try {
          await requestJSON("/api/v1/repos", {
            method: "POST",
            body: JSON.stringify({
              owner: form.get("owner"),
              owner_type: "organization",
              name: form.get("name"),
              description: form.get("description"),
              visibility: form.get("visibility"),
              default_branch: form.get("default_branch"),
            }),
          });
          window.location.reload();
        } catch (error) {
          showFlash(error.message);
        }
      });
    }
  };

  const renderOrganizations = async () => {
    const { organizations } = await requestJSON("/api/v1/orgs");
    const cards = organizations.length
      ? organizations
          .map(
            (organization) => `
              <article class="panel">
                <div class="button-row">
                  <span class="pill">${escapeHTML(organization.role)}</span>
                  <span class="pill">${escapeHTML(organization.organization_slug)}</span>
                </div>
                <h2>${escapeHTML(organization.organization_display_name)}</h2>
                <p class="muted">Add members directly from the browser when you are an organization owner.</p>
                <form class="form-stack" data-form="add-member" data-org="${escapeHTML(organization.organization_slug)}">
                  <label class="field"><span>Username</span><input name="username" required></label>
                  <label class="field">
                    <span>Role</span>
                    <select name="role">
                      <option value="member">Member</option>
                      <option value="maintainer">Maintainer</option>
                      <option value="owner">Owner</option>
                    </select>
                  </label>
                  <button class="button secondary" type="submit">Add Member</button>
                </form>
              </article>
            `
          )
          .join("")
      : `<div class="empty-state">You do not belong to any organizations yet.</div>`;

    app.innerHTML = `
      <section class="hero-panel">
        <p class="meta-text">Organizations</p>
        <h1>Shared ownership is now a first-class part of Forge.</h1>
        <p class="muted">Organization roles decide who can read, write, administer, and create repositories.</p>
      </section>
      <section class="stack">${cards}</section>
    `;

    app.querySelectorAll('[data-form="add-member"]').forEach((form) => {
      form.addEventListener("submit", async (event) => {
        event.preventDefault();
        const target = event.currentTarget;
        const org = target.dataset.org;
        const formData = new FormData(target);
        try {
          await requestJSON(`/api/v1/orgs/${encodeURIComponent(org)}/members`, {
            method: "POST",
            body: JSON.stringify({
              username: formData.get("username"),
              role: formData.get("role"),
            }),
          });
          showFlash(`Added ${formData.get("username")} to ${org}.`, "success");
          target.reset();
        } catch (error) {
          showFlash(error.message);
        }
      });
    });
  };

  const renderKeys = async () => {
    const [{ user }, { keys }] = await Promise.all([requestJSON("/api/v1/me"), requestJSON("/api/v1/keys")]);
    const keyCards = keys.length
      ? keys
          .map(
            (key) => `
              <article class="repo-card">
                <div class="button-row">
                  <span class="pill">${escapeHTML(key.name)}</span>
                  <span class="pill">${escapeHTML(key.fingerprint_sha256)}</span>
                </div>
                <p class="muted">Created ${escapeHTML(formatDate(key.created_at))}</p>
                <p class="muted">Last used ${escapeHTML(formatDate(key.last_used_at))}</p>
                <code class="code-block">${escapeHTML(key.public_key)}</code>
              </article>
            `
          )
          .join("")
      : `<div class="empty-state">No SSH keys registered for ${escapeHTML(user.username)} yet.</div>`;

    app.innerHTML = `
      <section class="hero-panel">
        <p class="meta-text">SSH Access</p>
        <h1>Register keys once, then use the same identity for browser, HTTP, and SSH workflows.</h1>
      </section>
      <section class="split-grid">
        <article class="panel">
          <h2>Add SSH Key</h2>
          <form class="form-stack" data-form="add-key">
            <label class="field"><span>Name</span><input name="name" required></label>
            <label class="field"><span>Public key</span><textarea name="public_key" required></textarea></label>
            <button class="button" type="submit">Save Key</button>
          </form>
        </article>
        <article class="panel">
          <h2>Registered Keys</h2>
          <div class="stack">${keyCards}</div>
        </article>
      </section>
    `;

    app.querySelector('[data-form="add-key"]').addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await requestJSON("/api/v1/keys", {
          method: "POST",
          body: JSON.stringify({
            name: form.get("name"),
            public_key: form.get("public_key"),
          }),
        });
        window.location.reload();
      } catch (error) {
        showFlash(error.message);
      }
    });
  };

  const renderRepository = async () => {
    const [detail, webhookPayload] = await Promise.all([
      requestJSON(`/api/v1/repos/${encodeURIComponent(repoOwner)}/${encodeURIComponent(repoName)}`),
      requestJSON(`/api/v1/repos/${encodeURIComponent(repoOwner)}/${encodeURIComponent(repoName)}/webhooks`).catch((error) => {
        if (error.status === 403) {
          return null;
        }
        throw error;
      }),
    ]);
    const repository = detail.repository;
    const webhooks = webhookPayload?.webhooks ?? null;
    const webhookCards =
      webhooks && webhooks.length
        ? webhooks
            .map(
              (webhook) => `
                <article class="repo-card">
                  <div class="button-row">
                    <span class="pill">${escapeHTML(webhook.events.join(", "))}</span>
                    <span class="pill">${escapeHTML(webhook.url)}</span>
                  </div>
                  <p class="muted">Successes: ${escapeHTML(webhook.success_count)} | Failures: ${escapeHTML(webhook.failure_count)}</p>
                  <p class="muted">Last delivery: ${escapeHTML(formatDate(webhook.last_delivery_at))}</p>
                  <p class="muted">${escapeHTML(webhook.last_delivery_error || "Last delivery recorded without error.")}</p>
                  <button class="button secondary" type="button" data-action="delete-webhook" data-webhook-id="${escapeHTML(webhook.id)}">Delete Webhook</button>
                </article>
              `
            )
            .join("")
        : `<div class="empty-state">No repository webhooks registered yet.</div>`;

    app.innerHTML = `
      <section class="hero-panel">
        <div class="button-row">
          <span class="pill">${escapeHTML(repository.owner_type)}</span>
          <span class="pill">${escapeHTML(repository.visibility)}</span>
        </div>
        <h1>${escapeHTML(repository.owner)}/${escapeHTML(repository.name)}</h1>
        <p class="muted">${escapeHTML(repository.description || "No description yet.")}</p>
      </section>
      <section class="split-grid">
        <article class="panel">
          <h2>Clone URLs</h2>
          <div class="stack">
            <div>
              <p class="meta-text">Smart HTTP</p>
              <code class="code-block">${escapeHTML(detail.http_clone_url)}</code>
            </div>
            ${
              detail.ssh_clone_url
                ? `<div>
                    <p class="meta-text">SSH</p>
                    <code class="code-block">${escapeHTML(detail.ssh_clone_url)}</code>
                  </div>`
                : ""
            }
          </div>
        </article>
        <article class="panel">
          <h2>Repository Status</h2>
          <div class="stack">
            <p class="muted">Default branch: <strong>${escapeHTML(repository.default_branch || "main")}</strong></p>
            <p class="muted">Size: <strong>${escapeHTML(repository.size_bytes || 0)} bytes</strong></p>
            <p class="muted">Last indexed: <strong>${escapeHTML(formatDate(repository.last_indexed_at))}</strong></p>
            <p class="muted">Last maintained: <strong>${escapeHTML(formatDate(repository.last_maintained_at))}</strong></p>
          </div>
        </article>
      </section>
      <section class="panel">
        <h2>Add Collaborator</h2>
        <form class="form-stack" data-form="add-collaborator">
          <label class="field"><span>Username</span><input name="username" required></label>
          <label class="field">
            <span>Role</span>
            <select name="role">
              <option value="read">Read</option>
              <option value="write">Write</option>
              <option value="admin">Admin</option>
            </select>
          </label>
          <button class="button secondary" type="submit">Add Collaborator</button>
        </form>
      </section>
      <section class="panel">
        <h2>Repository Webhooks</h2>
        ${
          webhooks === null
            ? `<p class="muted">Webhook management requires repository admin access.</p>`
            : `
              <form class="form-stack" data-form="add-webhook">
                <label class="field"><span>Delivery URL</span><input name="url" type="url" placeholder="https://example.com/hooks/forge" required></label>
                <label class="field"><span>Secret</span><input name="secret" placeholder="Optional signing secret"></label>
                <label class="field">
                  <span>Events</span>
                  <select name="event">
                    <option value="repository.push">repository.push</option>
                    <option value="repository.deleted">repository.deleted</option>
                  </select>
                </label>
                <button class="button" type="submit">Create Webhook</button>
              </form>
              <div class="stack">${webhookCards}</div>
            `
        }
      </section>
    `;

    app.querySelector('[data-form="add-collaborator"]').addEventListener("submit", async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      try {
        await requestJSON(`/api/v1/repos/${encodeURIComponent(repoOwner)}/${encodeURIComponent(repoName)}/collaborators`, {
          method: "POST",
          body: JSON.stringify({
            username: form.get("username"),
            role: form.get("role"),
          }),
        });
        showFlash(`Added ${form.get("username")} as a collaborator.`, "success");
        event.currentTarget.reset();
      } catch (error) {
        showFlash(error.message);
      }
    });

    const addWebhookForm = app.querySelector('[data-form="add-webhook"]');
    if (addWebhookForm) {
      addWebhookForm.addEventListener("submit", async (event) => {
        event.preventDefault();
        const form = new FormData(event.currentTarget);
        try {
          await requestJSON(`/api/v1/repos/${encodeURIComponent(repoOwner)}/${encodeURIComponent(repoName)}/webhooks`, {
            method: "POST",
            body: JSON.stringify({
              url: form.get("url"),
              secret: form.get("secret"),
              events: [form.get("event")],
            }),
          });
          window.location.reload();
        } catch (error) {
          showFlash(error.message);
        }
      });
    }

    app.querySelectorAll('[data-action="delete-webhook"]').forEach((button) => {
      button.addEventListener("click", async () => {
        try {
          await requestJSON(`/api/v1/repos/${encodeURIComponent(repoOwner)}/${encodeURIComponent(repoName)}/webhooks/${encodeURIComponent(button.dataset.webhookId)}`, {
            method: "DELETE",
          });
          window.location.reload();
        } catch (error) {
          showFlash(error.message);
        }
      });
    });
  };

  const boot = async () => {
    try {
      hydrateLogout();
      switch (view) {
        case "login":
          renderLogin();
          return;
        case "register":
          renderRegister();
          return;
        case "repos":
          await renderRepositories();
          return;
        case "orgs":
          await renderOrganizations();
          return;
        case "keys":
          await renderKeys();
          return;
        case "repo":
          await renderRepository();
          return;
        default:
          app.innerHTML = '<div class="empty-state">Unknown view.</div>';
      }
    } catch (error) {
      if (error.status === 401) {
        window.location.href = "/app/login";
        return;
      }
      showFlash(error.message || "Something went wrong while loading the page.");
    }
  };

  boot();
})();
