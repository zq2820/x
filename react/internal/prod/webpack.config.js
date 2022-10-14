const webpack = require("webpack");
const UnminifiedWebpackPlugin = require('unminified-webpack-plugin');

module.exports = {
	mode: 'development',
	entry: "./entry.js",
	output: {
		path: __dirname,
		filename: "prod.inc.js",
		libraryTarget: "this",
	},
	optimization: {
    minimize: false
  }
};
