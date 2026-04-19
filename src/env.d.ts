/// <reference types="astro/client" />

interface Window {
  __lightbox_initialized?: boolean;
  __lightbox_close?: () => void;
  __loupe_controller?: { detach: () => void };
  __dust_initialized?: boolean;
}

declare namespace App {
  // interface Locals {}
}
