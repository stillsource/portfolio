/// <reference types="astro/client" />

declare module '@fontsource/inter';
declare module '@fontsource/bodoni-moda';

interface Window {
  __lightbox_initialized?: boolean;
  __lightbox_close?: () => void;
  __loupe_controller?: { detach: () => void };
  __dust_initialized?: boolean;
}

declare namespace App {
  // interface Locals {}
}
