export default {
    'test': () => 'task test',
    'coverage': () => 'task coverage',
    'report': () => 'task report',
    'build': () => 'task build',
    'lint': () => 'putout . && task lint',
    'fix:lint': () => 'putout . --fix && task fix:lint',
};
