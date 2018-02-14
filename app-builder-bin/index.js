"use strict"

import {spawn as _spawn} from "child_process";

const path = require("path")

const nameMap = {
  "darwin": "mac",
  "win32": "win",
  "linux": "linux",
}

const suffix = nameMap[process.platform]
if (suffix == null) {
  throw new Error("Unsupported platform " + process.platform)
}

exports.appBuilderPath = process.env.USE_SYSTEM_APP_BUILDER === "true" ? "app-builder" : require(`app-builder-bin-${suffix}`).appBuilderPath

exports.execute = function (command, args) {
  return new Promise((resolve, reject) => {
    let env = getProcessEnv(null)
    if (options.stdio == null) {
      const isDebugEnabled = extraOptions == null || extraOptions.isDebugEnabled == null ? debug.enabled : extraOptions.isDebugEnabled
      // do not ignore stdout/stderr if not debug, because in this case we will read into buffer and print on error
      options.stdio = [extraOptions != null && extraOptions.isPipeInput ? "pipe" : "ignore", isDebugEnabled ? "inherit" : "pipe", isDebugEnabled ? "inherit" : "pipe"]
    }

    // use general debug.enabled to log spawn, because it doesn't produce a lot of output (the only line), but important in any case
    if (log.isDebugEnabled) {
      const argsString = args.join(" ")
      const logFields: any = {
        command,
        args: command === "docker" ? argsString : removePassword(argsString),
      }
      if (options != null && options.cwd != null) {
        logFields.cwd = options.cwd
      }
      log.debug(logFields, "spawning")
    }

    try {
      return _spawn(command, args, options)
    }
    catch (e) {
      throw new Error(`Cannot spawn ${command}: ${e.stack || e}`)
    }


    handleProcess("close", doSpawn(command, args || [], options, extraOptions), command, isCollectOutput, resolve, reject)
  })
}

function getProcessEnv(env) {
  if (process.platform === "win32") {
    return env
  }

  const finalEnv = {
    ...(env || process.env)
  }

  // without LC_CTYPE dpkg can returns encoded unicode symbols
  // set LC_CTYPE to avoid crash https://github.com/electron-userland/electron-builder/issues/503 Even "en_DE.UTF-8" leads to error.
  const locale = process.platform === "linux" ? (process.env.LANG || "C.UTF-8") : "en_US.UTF-8"
  finalEnv.LANG = locale
  finalEnv.LC_CTYPE = locale
  finalEnv.LC_ALL = locale
  return finalEnv
}