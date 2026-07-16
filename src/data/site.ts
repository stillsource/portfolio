// Single source of identity/contact. Change values here — hero, navbar, about
// and /collaborer all read from this object. Starting values are
// placeholders-that-work (real ik.me email, current personal Instagram); they
// migrate trivially to a domain email / dedicated Instagram later.
export const site = {
  name: 'Hors-Champ',
  role: 'Photographe de rue',
  // Double quotes: the copy contains an apostrophe (qu'on).
  tagline: "Je remets la lumière sur le beau qu'on ne regarde plus.",
  email: 'rodriguezalexandre8@ik.me',
  instagram: {
    handle: '@rzalexandre',
    url: 'https://instagram.com/rzalexandre',
  },
} as const;
