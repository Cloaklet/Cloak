// const FontminPlugin = require('fontmin-webpack')
// const NukeCssPlugin = require('nukecss-webpack')

module.exports = {
  productionSourceMap: false,

  chainWebpack: config => {
    // const fontsRule = config.module.rule('fonts')
    // fontsRule.uses.clear()
    // fontsRule
    //   .use('file-loader')
    //   .loader('file-loader')
    //   .end()

    // config.plugin('NukeCssPlugin')
    //   .use(NukeCssPlugin)

    // FontMin plugin greatly impacts the build time (+~300s),
    // disable it for now.
    // config.plugin('FontminPlugin')
    //   .use(FontminPlugin, [{autodetect: true, glyphs: []}])

  },

  pluginOptions: {
    i18n: {
      locale: 'en',
      fallbackLocale: 'en',
      localeDir: 'locales',
      enableInSFC: true
    }
  }
}
