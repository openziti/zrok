import React from 'react';
// Importing the original mapper + our components according to the Docusaurus doc
import MDXComponents from '@theme-original/MDXComponents';
import Card from '@zrokroot/src/components/Card';
import CardBody from '@zrokroot/src/components/Card/CardBody';
import CardFooter from '@zrokroot/src/components/Card/CardFooter';
import CardHeader from '@zrokroot/src/components/Card/CardHeader';
import CardImage from '@zrokroot/src/components/Card/CardImage';
import Columns from '@zrokroot/src/components/Columns';
import Column from '@zrokroot/src/components/Column';
export default {
  // Reusing the default mapping
  ...MDXComponents,
  Card,
  CardHeader,
  CardBody,
  CardFooter,
  CardImage,
  Columns,
  Column,
};