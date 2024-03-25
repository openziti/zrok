import React from 'react';
import yaml from 'js-yaml';
import CodeBlock from '@theme/CodeBlock';

const ConcatenateYamlSnippets = ({ title, children }) => {

  // Convert each YAML object to a string and concatenate
  const concatenatedYaml = children.map(child => {
    // Check if the child is a string or an object
    if (typeof child === 'string') {
      // If it's a string, use it as is
      return child.trim();
    } else {
      // If it's an object, convert it to a YAML string
      return yaml.dump(child).trim();
    }
  }).join('\n\n');

  return (
    <div>
        <CodeBlock
          language="yaml"
          title={title}
        >
          {concatenatedYaml}
        </CodeBlock>
    </div>
  );
};

export default ConcatenateYamlSnippets;
