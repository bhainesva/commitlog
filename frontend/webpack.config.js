import HtmlWebpackPlugin from 'html-webpack-plugin';

const config = {
    mode: 'development',
    entry: './src/index.js',
    devServer: {
      contentBase: './dist',
    },
    output: {
        filename: 'bundle.js',
        path: new URL('./dist', import.meta.url).pathname,
        clean: true,
    },
    plugins: [
        new HtmlWebpackPlugin({
            title: 'Commit Logs',
        }),
    ],
    module: {
        rules: [
            {
                test: /\.css$/i,
                use: ['style-loader', 'css-loader'],
            },
            {
                test: /\.(png|svg|jpg|jpeg|gif)$/i,
                type: 'asset/resource',
            },
        ],
    },
}

export default config;