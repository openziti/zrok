const ModuleScopePlugin = require("react-dev-utils/ModuleScopePlugin");
const path = require("path");

module.exports = function override(config, env) {
    //do stuff with the webpack config...

    config.resolve.plugins.forEach(plugin => {
        if (plugin instanceof ModuleScopePlugin) {
          plugin.allowedFiles.add(path.resolve("./node_modules/querystring-es3/index.js"));
        }
    });

    let loaders = config.resolve
    loaders.fallback = {
        "querystring": require.resolve("querystring-es3")
    }
    return config;
}
