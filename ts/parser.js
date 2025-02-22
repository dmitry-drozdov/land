"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var parse = require("@typescript-eslint/parser").parse;
var fs = require("fs");
var path = require("path");
// 📂 Папка с TypeScript-файлами (исходная, где лежат `.ts` файлы)
var SOURCE_DIR = "e:/phd/ts";
// 📂 Папка для сохранения `.res` файлов (результатов)
var OUTPUT_DIR = "e:/phd/ts_res";
//  🔄 Функция для поиска резолверов в AST
function findResolvers(ast) {
    var resolvers = [];
    function traverse(node, className, factoryName) {
        var _a;
        if (node.type === "Property") {
            var keyName = node.key.name || node.key.value;
            if (node.value.type === "ArrowFunctionExpression" || node.value.type === "FunctionExpression") {
                // Старый способ:
                // resolvers.push(factoryName ? `${factoryName}.${keyName}` : keyName);
                // Новый способ: только имя функции
                resolvers.push(keyName);
            }
        }
        if (node.type === "MethodDefinition" && node.key.type === "Identifier") {
            // Старый способ:
            // resolvers.push(`${className}.${node.key.name}`);
            // Новый способ: только имя функции
            resolvers.push(node.key.name);
        }
        if (node.type === "ClassDeclaration" && node.id) {
            className = node.id.name;
        }
        if (node.type === "VariableDeclarator" && ((_a = node.init) === null || _a === void 0 ? void 0 : _a.type) === "ArrowFunctionExpression") {
            factoryName = node.id.name;
        }
        for (var key in node) {
            if (node[key] && typeof node[key] === "object") {
                traverse(node[key], className, factoryName);
            }
        }
    }
    traverse(ast);
    return resolvers;
}
// 🔍 Функция для обработки одного файла
function processFile(filePath) {
    try {
        var code = fs.readFileSync(filePath, "utf-8");
        var ast = parse(code, {
            ecmaVersion: "latest",
            sourceType: "module",
            range: true,
            loc: true
        });
        var resolvers = findResolvers(ast);
        if (resolvers.length > 0) {
            // 📂 Сохраняем файлы в аналогичную структуру внутри OUTPUT_DIR
            var relativePath = path.relative(SOURCE_DIR, filePath);
            var resFilePath = path.join(OUTPUT_DIR, relativePath.replace(/\.ts$/, ".res"));
            // 🛠 Создаем вложенные директории, если их нет
            fs.mkdirSync(path.dirname(resFilePath), { recursive: true });
            // 💾 Записываем результат
            fs.writeFileSync(resFilePath, resolvers.join("\n"), "utf-8");
            console.log("\u2705 \u0420\u0435\u0437\u043E\u043B\u0432\u0435\u0440\u044B \u0441\u043E\u0445\u0440\u0430\u043D\u0435\u043D\u044B \u0432: ".concat(resFilePath));
        }
    }
    catch (err) {
        console.error("\u274C \u041E\u0448\u0438\u0431\u043A\u0430 \u043F\u0440\u0438 \u043E\u0431\u0440\u0430\u0431\u043E\u0442\u043A\u0435 \u0444\u0430\u0439\u043B\u0430 ".concat(filePath, ":"), err);
    }
}
// 📂 Рекурсивный обход папки и обработка `.ts` файлов
function processDirectory(directory) {
    var files = fs.readdirSync(directory);
    files.forEach(function (file) {
        var fullPath = path.join(directory, file);
        var stat = fs.statSync(fullPath);
        if (stat.isDirectory()) {
            processDirectory(fullPath);
        }
        else if (file.endsWith(".ts")) {
            console.log("\uD83D\uDD0D \u041E\u0431\u0440\u0430\u0431\u0430\u0442\u044B\u0432\u0430\u0435\u043C \u0444\u0430\u0439\u043B: ".concat(fullPath));
            processFile(fullPath);
        }
    });
}
// 🚀 Запуск
console.log("\uD83D\uDCC2 \u041D\u0430\u0447\u0438\u043D\u0430\u0435\u043C \u0430\u043D\u0430\u043B\u0438\u0437 \u0432 \u043F\u0430\u043F\u043A\u0435: ".concat(SOURCE_DIR));
console.log("\uD83D\uDCC2 \u0420\u0435\u0437\u0443\u043B\u044C\u0442\u0430\u0442\u044B \u0441\u043E\u0445\u0440\u0430\u043D\u044F\u0435\u043C \u0432: ".concat(OUTPUT_DIR));
processDirectory(SOURCE_DIR);
console.log("✅ Анализ завершен.");
