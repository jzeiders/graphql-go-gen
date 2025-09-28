#!/usr/bin/env node

const fs = require('fs').promises;
const path = require('path');
const diff = require('diff');
const { execSync } = require('child_process');

// Test cases mapping
const testCases = [
  {
    name: 'default',
    config: 'codegen.default.ts',
    jsFile: './generated/default.ts',
    goFile: './go-generated/default.ts',
    goTestData: '../../pkg/plugins/typescript_operations/testdata/default.ts'
  },
  {
    name: 'immutable',
    config: 'codegen.immutable.ts',
    jsFile: './generated/immutable.ts',
    goFile: './go-generated/immutable.ts',
    goTestData: '../../pkg/plugins/typescript_operations/testdata/immutable.ts'
  },
  {
    name: 'skip-typename',
    config: 'codegen.skip-typename.ts',
    jsFile: './generated/skip-typename.ts',
    goFile: './go-generated/skip-typename.ts',
    goTestData: '../../pkg/plugins/typescript_operations/testdata/skip-typename.ts'
  },
  {
    name: 'omit-suffix',
    config: 'codegen.omit-suffix.ts',
    jsFile: './generated/omit-suffix.ts',
    goFile: './go-generated/omit-suffix.ts',
    goTestData: '../../pkg/plugins/typescript_operations/testdata/omit-suffix.ts'
  },
  {
    name: 'flatten',
    config: 'codegen.flatten.ts',
    jsFile: './generated/flatten.ts',
    goFile: './go-generated/flatten.ts',
    goTestData: '../../pkg/plugins/typescript_operations/testdata/flatten.ts'
  },
  {
    name: 'avoid-optionals',
    config: 'codegen.avoid-optionals.ts',
    jsFile: './generated/avoid-optionals.ts',
    goFile: './go-generated/avoid-optionals.ts',
    goTestData: '../../pkg/plugins/typescript_operations/testdata/avoid-optionals.ts'
  }
];

// Color codes for console output
const colors = {
  reset: '\x1b[0m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  gray: '\x1b[90m'
};

// Normalization function to handle minor differences
function normalizeContent(content) {
  return content
    // Remove trailing whitespace
    .replace(/[ \t]+$/gm, '')
    // Normalize line endings
    .replace(/\r\n/g, '\n')
    // Remove empty lines at the end
    .replace(/\n+$/, '\n');
}

async function runGoGenerator(testCase) {
  const binaryPath = '../../graphql-go-gen';

  try {
    // Create output directory
    await fs.mkdir('go-generated', { recursive: true });

    // Run graphql-go-gen
    console.log(`${colors.gray}  Running: graphql-go-gen generate -c ${testCase.config}${colors.reset}`);
    execSync(`${binaryPath} generate -c ${testCase.config}`, {
      stdio: 'pipe',
      encoding: 'utf8'
    });

    return { success: true };
  } catch (error) {
    return { success: false, error: error.message };
  }
}

async function compareThreeWay(testCase) {
  const results = {
    testCase,
    comparisons: {}
  };

  try {
    // Read all three files
    const [jsContent, goContent, goTestDataContent] = await Promise.all([
      fs.readFile(testCase.jsFile, 'utf8').catch(() => null),
      fs.readFile(testCase.goFile, 'utf8').catch(() => null),
      fs.readFile(testCase.goTestData, 'utf8').catch(() => null)
    ]);

    // Normalize content
    const normalizedJs = jsContent ? normalizeContent(jsContent) : null;
    const normalizedGo = goContent ? normalizeContent(goContent) : null;
    const normalizedTestData = goTestDataContent ? normalizeContent(goTestDataContent) : null;

    // Compare Go Generated vs Go TestData (expected)
    if (normalizedGo && normalizedTestData) {
      results.comparisons.goVsExpected = {
        matches: normalizedGo === normalizedTestData,
        changes: normalizedGo !== normalizedTestData ?
          diff.diffLines(normalizedTestData, normalizedGo) : null
      };
    }

    // Compare JS Generated vs Go Generated
    if (normalizedJs && normalizedGo) {
      results.comparisons.jsVsGo = {
        matches: normalizedJs === normalizedGo,
        changes: normalizedJs !== normalizedGo ?
          diff.diffLines(normalizedGo, normalizedJs) : null
      };
    }

    // Compare JS Generated vs Go TestData (for reference)
    if (normalizedJs && normalizedTestData) {
      results.comparisons.jsVsExpected = {
        matches: normalizedJs === normalizedTestData,
        changes: normalizedJs !== normalizedTestData ?
          diff.diffLines(normalizedTestData, normalizedJs) : null
      };
    }

    results.success = true;
  } catch (error) {
    results.success = false;
    results.error = error.message;
  }

  return results;
}

async function generateReport(results) {
  const report = [];

  report.push('# GraphQL Codegen Parity Test Report\n');
  report.push(`Test Date: ${new Date().toISOString()}\n`);
  report.push('\n## Summary\n');

  // Count successes
  let configCompatible = 0;
  let goMatchesExpected = 0;
  let jsMatchesGo = 0;
  let totalTests = results.length;

  for (const result of results) {
    if (result.goGenerated?.success) configCompatible++;
    if (result.comparisons?.goVsExpected?.matches) goMatchesExpected++;
    if (result.comparisons?.jsVsGo?.matches) jsMatchesGo++;
  }

  report.push(`### Configuration Compatibility`);
  report.push(`- Go can read TypeScript configs: ${configCompatible}/${totalTests}`);
  report.push(`- Go output matches expected testdata: ${goMatchesExpected}/${totalTests}`);
  report.push(`- JS output matches Go output: ${jsMatchesGo}/${totalTests}`);
  report.push('');

  // Detailed results for each test
  report.push('\n## Detailed Results\n');

  for (const result of results) {
    report.push(`\n### ${result.testCase.name}\n`);
    report.push(`Config: \`${result.testCase.config}\`\n`);

    // Config compatibility
    if (result.goGenerated?.success) {
      report.push('✅ Go successfully consumed TypeScript config');
    } else {
      report.push(`❌ Go failed to consume TypeScript config: ${result.goGenerated?.error || 'Unknown error'}`);
    }

    // Comparison results
    if (result.comparisons) {
      report.push('\n#### Comparisons:');

      // Go vs Expected
      if (result.comparisons.goVsExpected) {
        if (result.comparisons.goVsExpected.matches) {
          report.push('- ✅ Go output matches expected testdata');
        } else {
          report.push('- ❌ Go output differs from expected testdata');
          if (result.comparisons.goVsExpected.changes) {
            report.push('\n<details>');
            report.push('<summary>Differences (Go Generated vs Expected)</summary>\n');
            report.push('```diff');

            let lineCount = 0;
            const maxLines = 30;
            for (const part of result.comparisons.goVsExpected.changes) {
              if (lineCount >= maxLines) {
                report.push('... (diff truncated)');
                break;
              }
              const lines = part.value.split('\n').filter(line => line);
              for (const line of lines.slice(0, maxLines - lineCount)) {
                if (part.added) {
                  report.push(`+ ${line}`);
                } else if (part.removed) {
                  report.push(`- ${line}`);
                }
                lineCount++;
              }
            }
            report.push('```');
            report.push('</details>\n');
          }
        }
      }

      // JS vs Go
      if (result.comparisons.jsVsGo) {
        if (result.comparisons.jsVsGo.matches) {
          report.push('- ✅ JS output matches Go output (full parity!)');
        } else {
          report.push('- ⚠️  JS output differs from Go output');
          if (result.comparisons.jsVsGo.changes) {
            const adds = result.comparisons.jsVsGo.changes.filter(c => c.added).length;
            const removes = result.comparisons.jsVsGo.changes.filter(c => c.removed).length;
            report.push(`  Differences: +${adds} sections, -${removes} sections`);
          }
        }
      }
    }
  }

  // Configuration mapping reference
  report.push('\n## Configuration Mapping\n');
  report.push('| Test Case | Config Option | Value |');
  report.push('|-----------|---------------|-------|');
  report.push('| default | - | default settings |');
  report.push('| immutable | immutableTypes | true |');
  report.push('| skip-typename | skipTypename | true |');
  report.push('| omit-suffix | omitOperationSuffix | true |');
  report.push('| flatten | flattenGeneratedTypes | true |');
  report.push('| avoid-optionals | avoidOptionals | true |');

  report.push('\n## Legend\n');
  report.push('- ✅ Full match');
  report.push('- ⚠️  Partial match or minor differences');
  report.push('- ❌ Significant differences or failure');

  await fs.writeFile('parity-report.md', report.join('\n'));
  return report.join('\n');
}

async function main() {
  console.log(`${colors.blue}🔍 GraphQL Codegen Parity Test${colors.reset}\n`);

  // Step 1: Generate JS outputs
  console.log(`${colors.blue}Step 1: Generating JavaScript outputs...${colors.reset}`);
  try {
    execSync('npm run codegen:all', { stdio: 'inherit' });
  } catch (error) {
    console.error(`${colors.red}Failed to generate JS outputs${colors.reset}`);
    process.exit(1);
  }

  // Step 2: Generate Go outputs for each config
  console.log(`\n${colors.blue}Step 2: Generating Go outputs...${colors.reset}`);
  const goResults = [];
  for (const testCase of testCases) {
    const result = await runGoGenerator(testCase);
    goResults.push(result);
    if (result.success) {
      console.log(`${colors.green}✓ ${testCase.name}${colors.reset}`);
    } else {
      console.log(`${colors.red}✗ ${testCase.name}: ${result.error}${colors.reset}`);
    }
  }

  // Step 3: Compare outputs
  console.log(`\n${colors.blue}Step 3: Comparing outputs...${colors.reset}`);
  const results = [];
  for (let i = 0; i < testCases.length; i++) {
    const comparison = await compareThreeWay(testCases[i]);
    comparison.goGenerated = goResults[i];
    results.push(comparison);

    // Print summary
    const goVsExp = comparison.comparisons?.goVsExpected?.matches;
    const jsVsGo = comparison.comparisons?.jsVsGo?.matches;

    let status = colors.green + '✅';
    if (!goVsExp || !jsVsGo) {
      status = colors.yellow + '⚠️';
    }
    if (!goResults[i].success) {
      status = colors.red + '❌';
    }

    console.log(`${status} ${testCases[i].name}${colors.reset}`);
    if (goVsExp !== undefined) {
      console.log(`${colors.gray}   Go vs Expected: ${goVsExp ? '✓' : '✗'}${colors.reset}`);
    }
    if (jsVsGo !== undefined) {
      console.log(`${colors.gray}   JS vs Go: ${jsVsGo ? '✓' : '✗'}${colors.reset}`);
    }
  }

  // Step 4: Generate report
  console.log(`\n${colors.blue}Step 4: Generating report...${colors.reset}`);
  await generateReport(results);
  console.log(`${colors.green}✅ Report saved to parity-report.md${colors.reset}`);

  // Exit code based on results
  const hasFailures = results.some(r =>
    !r.goGenerated?.success ||
    !r.comparisons?.goVsExpected?.matches
  );

  if (hasFailures) {
    console.log(`\n${colors.yellow}⚠️  Some tests failed. See parity-report.md for details.${colors.reset}`);
    process.exit(1);
  } else {
    console.log(`\n${colors.green}🎉 All tests passed!${colors.reset}`);
  }
}

main().catch(error => {
  console.error(`${colors.red}Fatal error: ${error}${colors.reset}`);
  process.exit(1);
});