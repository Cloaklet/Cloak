# Cloak

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

### Compiles and minifies for production
```
npm run build
```

### Lints and fixes files
```
npm run lint
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

## I18N

Locale files reside in `src/locales`, they are JSON files. Key paths are used to identify the strings to be translated, so be sure not to change the key when translating them.

Also there are some special keys:

- `zxcvbn`: sub-keys are taken from strings from [this file](https://github.com/dropbox/zxcvbn/blob/67c4ece9efc40c9d0a1d7d995b2b22a91be500c2/src/feedback.coffee), because zxcvbn does not support i18n so we're dealing with it ourselves.