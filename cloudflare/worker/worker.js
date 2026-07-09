// Routes blog API traffic to the phone if it's healthy, otherwise falls back to Render.
// PHONE_URL is set in the Cloudflare dashboard (changes often with Quick Tunnels).
// RENDER_URL is set in wrangler.toml (stable).

export default {
  async fetch(request, env) {
    const url = new URL(request.url);

    if (env.PHONE_URL) {
      try {
        const controller = new AbortController();
        const timeout = setTimeout(() => controller.abort(), 3000);
        const health = await fetch(env.PHONE_URL + "/health", {
          signal: controller.signal,
        });
        clearTimeout(timeout);

        if (health.ok) {
          const target = env.PHONE_URL + url.pathname + url.search;
          return fetch(new Request(target, request));
        }
      } catch {
        // phone unreachable or timed out, fall through to Render
      }
    }

    const target = env.RENDER_URL + url.pathname + url.search;
    return fetch(new Request(target, request));
  },
};
