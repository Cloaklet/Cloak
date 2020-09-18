const FontminPlugin = require('fontmin-webpack')

module.exports = {
  productionSourceMap: false,

  chainWebpack: config => {
    const fontsRule = config.module.rule('fonts')
    fontsRule.uses.clear()
    fontsRule
      .use('file-loader')
      .loader('file-loader')
    config.plugin('FontminPlugin')
      .use(FontminPlugin, [{autodetect: true, glyphs: []}])
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
