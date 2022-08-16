export function isCollapsable(arg: any): boolean {
    return arg instanceof Object && Object.keys(arg).length > 0;
}
export function isUrl(string: string): boolean {
    var urlRegexp = /^(https?:\/\/|ftps?:\/\/)?([a-z0-9%-]+\.){1,}([a-z0-9-]+)?(:(\d{1,5}))?(\/([a-z0-9\-._~:/?#[\]@!$&'()*+,;=%]+)?)?$/i;
    return urlRegexp.test(string);
}
export function json2html(json: string | any, options: any): string {
    var html = '';
    if (typeof json === 'string') {
        // Escape tags and quotes
        json = json
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/'/g, '&apos;')
            .replace(/"/g, '&quot;');

        if (options.withLinks && isUrl(json)) {
            html += '<a href="javascript:;" class="json-string">' + json + '</a>';
        } else {
            // Escape double quotes in the rendered non-URL string.
            json = json.replace(/&quot;/g, '\\&quot;');
            html += '<span class="json-string">"' + json + '"</span>';
        }
    } else if (typeof json === 'number') {
        html += '<span class="json-literal">' + json + '</span>';
    } else if (typeof json === 'boolean') {
        html += '<span class="json-literal">' + json + '</span>';
    } else if (json === null) {
        html += '<span class="json-literal">null</span>';
    } else if (json instanceof Array) {
        if (json.length > 0) {
            html += '[<ol class="json-array">';
            for (var i = 0; i < json.length; ++i) {
                html += '<li>';
                // Add toggle button if item is collapsable
                if (isCollapsable(json[i])) {
                    html += '<a href class="json-toggle"></a>';
                }
                html += json2html(json[i], options);
                // Add comma if item is not last
                if (i < json.length - 1) {
                    html += ',';
                }
                html += '</li>';
            }
            html += '</ol>]';
        } else {
            html += '[]';
        }
    } else if (typeof json === 'object') {
        var keyCount = Object.keys(json).length;
        if (keyCount > 0) {
            html += '{<ul class="json-dict">';
            for (var key in json) {
                if (Object.prototype.hasOwnProperty.call(json, key)) {
                    html += '<li>';
                    var keyRepr = options.withQuotes ?
                        '<span class="json-string">"' + key + '"</span>' : key;
                    // Add toggle button if item is collapsable
                    if (isCollapsable(json[key])) {
                        html += '<a href class="json-toggle">' + keyRepr + '</a>';
                    } else {
                        html += keyRepr;
                    }
                    html += ': ' + json2html(json[key], options);
                    // Add comma if item is not last
                    if (--keyCount > 0) {
                        html += ',';
                    }
                    html += '</li>';
                }
            }
            html += '</ul>}';
        } else {
            html += '{}';
        }
    }
    return html;
}
export function valueSplice(arr: string[], obj: any) {
    for (const key in obj) {
        if (key === arr[0]) {
            obj[key] = arr[1]
            break
        }
        if (Object.prototype.toString.call(obj[key]).slice(8) === 'Object]') {
            const newObj = obj[key]
            valueSplice(arr, newObj)
        }
    }
}
