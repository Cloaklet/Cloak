# Cloak

# frontend

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VSCode](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Type Support for `.vue` Imports in TS

TypeScript cannot handle type information for `.vue` imports by default, so we replace the `tsc` CLI with `vue-tsc` for type checking. In editors, we need [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) to make the TypeScript language service aware of `.vue` types.

## Customize configuration

See [Vite Configuration Reference](https://vitejs.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Type-Check, Compile and Minify for Production

```sh
npm run build
```

### Run Unit Tests with [Vitest](https://vitest.dev/)

```sh
npm run test:unit
```

### Lint with [ESLint](https://eslint.org/)

```sh
npm run lint
```

## I18N

Locale files reside in `src/locales`, they are JSON files. Key paths are used to identify the strings to be translated, so be sure not to change the key when translating them.

Also there are some special keys:

- `zxcvbn`: sub-keys are taken from strings from [this file](https://github.com/dropbox/zxcvbn/blob/67c4ece9efc40c9d0a1d7d995b2b22a91be500c2/src/feedback.coffee), because zxcvbn does not support i18n so we're dealing with it ourselves.
