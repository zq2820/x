const webpack = require("webpack");

module.exports = {
	mode: 'production',
	entry: "./entry.js",
	output: {
		path: __dirname,
		filename: "prod.inc.js",
		libraryTarget: "this",
	},
	optimization: {
    minimize: true
  }
};
