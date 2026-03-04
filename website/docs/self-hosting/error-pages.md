---
title: Error Pages
sidebar_label: Error Pages
sidebar_position: 17
---

# Custom Error Pages

zrok includes a built-in error page template that displays user-friendly messages for various error conditions like "share not found", "unauthorized access", and "bad gateway" errors. This template can be replaced with your own custom HTML file to match your organization's branding or provide custom error handling.

## Overview

The error page system uses Go's `text/template` package to render HTML pages with dynamic content. The template receives data through a `VariableData` struct containing:

- `Title`: Page title (appears in browser tab)
- `Banner`: Main heading text
- `Message`: Optional explanatory message
- `Error`: Optional error details

## Configuration Options

### Private Access (`zrok2 access private`)

For private access frontends, use the `--template-path` flag:

```bash
zrok2 access private --template-path /path/to/custom-template.html <shareToken>
```

### Public Frontend (`zrok2 access public`)

For public frontends, add the `template_path` configuration option to your frontend configuration YAML:

```yaml
v: 4
identity: public
address: 0.0.0.0:8080

# Path to custom error page template
template_path: /path/to/custom-template.html

# Other configuration options...
```

Then start the public frontend:

```bash
zrok2 access public /path/to/frontend-config.yml
```

## Template Structure

The template uses Go's template syntax with the following available variables:

- `{{.Title}}`: Page title
- `{{.Banner}}`: Main heading (may contain HTML)
- `{{.Message}}`: Optional message (may contain HTML)
- `{{.Error}}`: Optional error object

### Conditional Content

Use Go template conditionals to show content only when data is available:

```html
{{if .Message}}
<p>{{.Message}}</p>
{{end}}

{{if .Error}}
<div class="error-box">
    <strong>Error:</strong> {{.Error}}
</div>
{{end}}
```

## Custom Template Example

Here's a simplified version of the default template that you can customize:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <title>zrok - {{.Title}}</title>
    <style>
        body {
            margin: 0;
            padding: 25px;
            font-family: 'Arial', sans-serif;
            background-color: #f0f0f0;
            color: #333;
        }
        
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 20px;
            text-align: center;
        }
        
        .header h1 {
            margin: 0;
            font-size: 2.5rem;
            font-weight: bold;
        }
        
        .content {
            padding: 40px 20px;
            text-align: center;
        }
        
        .banner {
            font-size: 1.5rem;
            margin-bottom: 20px;
            color: #333;
        }
        
        .message {
            font-size: 1.1rem;
            line-height: 1.6;
            color: #666;
            margin-bottom: 20px;
        }
        
        .error {
            background-color: #fee;
            border: 1px solid #fcc;
            border-radius: 4px;
            padding: 15px;
            margin: 20px 0;
            font-family: 'Courier New', monospace;
            text-align: left;
        }
        
        .error strong {
            color: #d00;
        }
        
        @media (max-width: 600px) {
            body {
                padding: 10px;
            }
            
            .header h1 {
                font-size: 1.8rem;
            }
            
            .content {
                padding: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Your Service</h1>
        </div>
        
        <div class="content">
            <div class="banner">{{.Banner}}</div>
            
            {{if .Message}}
            <div class="message">{{.Message}}</div>
            {{end}}
            
            {{if .Error}}
            <div class="error">
                <strong>Error:</strong> {{.Error}}
            </div>
            {{end}}
            
            <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
            
            <p style="color: #999; font-size: 0.9rem;">
                Powered by <a href="https://zrok.io" style="color: #667eea;">zrok</a>
            </p>
        </div>
    </div>
</body>
</html>
```

## Error Types

Your template will be used for various error conditions:

### Share Not Found (404)
- **Title**: `'<shareToken>' not found!`
- **Banner**: `share <code><shareToken></code> not found!`
- **Message**: `are you running <code>zrok2 share</code> for this share?`

### Unauthorized Access (401)
- **Title**: `unauthorized!`
- **Banner**: `user not authorized!` or `<code><username></code> not authorized!`

### Bad Gateway (502)
- **Title**: Custom title based on the error
- **Banner**: Custom banner based on the error
- **Error**: Detailed error information

### Health Check (200)
- **Title**: `healthy`
- **Banner**: `healthy`

## Best Practices

1. **Keep it simple**: Error pages should load quickly and not depend on external resources that might also be failing.

2. **Responsive design**: Ensure your template works well on mobile devices.

3. **Clear messaging**: Provide helpful information to users about what went wrong and what they can do.

4. **Consistent branding**: Match your organization's visual identity.

5. **Escape user content**: Be cautious with user-provided content. The `Banner` and `Message` fields may contain HTML from the application.

6. **Test thoroughly**: Test your template with different error conditions to ensure it renders correctly.

7. **Fallback styling**: Include all CSS inline or use web fonts with fallbacks to ensure the page displays correctly even if external resources fail.
