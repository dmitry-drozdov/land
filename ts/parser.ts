const { parse } = require("@typescript-eslint/parser");
import * as fs from "fs";
import * as path from "path";

// 📂 Папка с TypeScript-файлами (исходная, где лежат `.ts` файлы)
const SOURCE_DIR = "e:/phd/ts";

// 📂 Папка для сохранения `.res` файлов (результатов)
const OUTPUT_DIR = "e:/phd/ts_res";

//  🔄 Функция для поиска резолверов в AST
function findResolvers(ast: any) {
    const resolvers: string[] = [];

    function traverse(node: any, className?: string, factoryName?: string, propertyDepth = 0) {
        //console.log(node);
        if (node.type === "ExportDefaultDeclaration" && node.declaration.type === "ObjectExpression") {
            return;
        }

        //  Исключаем деструктурирующее присваивание
        if (node.type === "VariableDeclarator" && node.id.type === "ObjectPattern") {
            return; // Не добавляем деструктурированные переменные
        }

        if (node.type === "FunctionDeclaration" || node.type === "TSTypeAliasDeclaration"){
            return;
        }

        if (node.type === "CallExpression") {
            return;
        }

        if (node.type === "ObjectPattern") {
            return;
        }

        if (node.type === "NewExpression") {
            return;
        }

        // Исключаем объекты с динамическими ключами (например, `[COLD_WATER_METER_RESOURCE_ID]: SnowflakeIcon`)
        if (node.type === "Property" && node.computed) {
            return;
        }

        if (node.type === "ArrowFunctionExpression") {
            for (const key in node) {
                if (node[key] && (typeof node[key] === "object") && node[key].type === "ObjectExpression") {
                    traverse(node[key], className, factoryName, propertyDepth);
                }
            }
            return;
        }

        if (node.type === "Property") {
            if (node.value.type === "ArrayExpression" || node.value.type === "Literal") {
                return;
            }

            const keyName = node.key.name || node.key.value;
            if (node.value.type === "ArrowFunctionExpression" || node.value.type === "FunctionExpression") {
                if (propertyDepth == 1)
                    resolvers.push(keyName);
                return;
            }

            if (node.value.type === "Identifier" || node.value.type === "MemberExpression" || node.value.type === "ConditionalExpression") {
                if (propertyDepth == 1)
                    resolvers.push(keyName);
                return;
            }
            propertyDepth++;
        }

        // if (node.type)
        //     console.log(node.type, node.key?.name || node.key?.value);

        if (node.type === "MethodDefinition" && node.key.type === "Identifier") {
            if (node.value.body) { // Исключаем методы без тела
                resolvers.push(node.key.name);
            }
            return;
        }

        if (node.type === "ClassDeclaration" && node.id) {
            className = node.id.name;
        }

        if (node.type === "VariableDeclarator" && node.init?.type === "ArrowFunctionExpression") {
            factoryName = node.id.name;
        }


        for (const key in node) {
            if (node[key] && typeof node[key] === "object") {
                traverse(node[key], className, factoryName, propertyDepth);
            }
        }
    }

    traverse(ast);
    return resolvers;
}

// 🔍 Функция для обработки одного файла
function processFile(filePath: string) {
    try {
        const code = fs.readFileSync(filePath, "utf-8");
        const ast = parse(code, {
            ecmaVersion: "latest",
            sourceType: "module",
            range: true,
            loc: true
        });

        const resolvers = findResolvers(ast);
        //console.log(resolvers);
        if (resolvers.length > 0) {
            // 📂 Сохраняем файлы в аналогичную структуру внутри OUTPUT_DIR
            const relativePath = path.relative(SOURCE_DIR, filePath);
            const resFilePath = path.join(OUTPUT_DIR, relativePath.replace(/\.ts$/, ".res"));

            // 🛠 Создаем вложенные директории, если их нет
            fs.mkdirSync(path.dirname(resFilePath), { recursive: true });

            // 💾 Записываем результат
            fs.writeFileSync(resFilePath, resolvers.join("\n"), "utf-8");
            //console.log(`✅ Резолверы сохранены в: ${resFilePath}`);
        }
    } catch (err) {
        console.error(`❌ Ошибка при обработке файла ${filePath}:`, err);
    }
}

// 📂 Рекурсивный обход папки и обработка `.ts` файлов
function processDirectory(directory: string) {
    const files = fs.readdirSync(directory);
    files.forEach(file => {
        const fullPath = path.join(directory, file);
        const stat = fs.statSync(fullPath);

        if (stat.isDirectory()) {
            processDirectory(fullPath);
        } else if (file.endsWith(".ts")) {
            //console.log(`🔍 Обрабатываем файл: ${fullPath}`);
            processFile(fullPath);
        }
    });
}

// 🚀 Запуск
console.log(`📂 Начинаем анализ в папке: ${SOURCE_DIR}`);
console.log(`📂 Результаты сохраняем в: ${OUTPUT_DIR}`);
processDirectory(SOURCE_DIR);
console.log("✅ Анализ завершен.");