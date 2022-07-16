const webpack = require("webpack");

module.exports = {
	mode: 'development',
  entry: "./entry.js",
  output: {
		path: __dirname,
    filename: "dev.inc.js",
    libraryTarget: "this",
  },
  devServer: {
    port: 8888
  }
};
