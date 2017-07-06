const path = require('path')

const include = path.join(__dirname, 'src')

module.exports = {
  entry: './src/index.js',
  output: {
    path: path.join(__dirname, 'dist'),
    filename: 'sacrifical-socket.js',
    libraryTarget: 'umd',
    library: 'sacrificial-socket',
  },
  devtool: 'source-map',
  module: {
    loaders: [
      {test: /\.js$/, loader: 'babel-loader', include}
    ]
  }
}