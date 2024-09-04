  import React,  { CSSProperties } from 'react'; // CSSProperties allows inline styling with better type checking.
  import clsx from 'clsx'; // clsx helps manage conditional className names in a clean and concise manner.
  const CardHeader = ({
    className, // classNamees for the container card
    style, // Custom styles for the container card
    children, // Content to be included within the card
    textAlign, 
    variant, 
    italic = false , 
    noDecoration = false, 
    transform, 
    breakWord = false, 
    truncate = false, 
    weight,
  }) => {   
    const text = textAlign ? `text--${textAlign}` :'';
    const textColor = variant ? `text--${variant}` : '';
    const textItalic = italic ? 'text--italic' : '';
    const textDecoration = noDecoration ? 'text-no-decoration' : '';
    const textType = transform ? `text--${transform}` : '';
    const textBreak = breakWord ? 'text--break' : '';
    const textTruncate = truncate ? 'text--truncate' : '';
    const textWeight = weight ? `text--${weight}` : '';
    return (
      <div
        className={clsx(
          'card__header',
          className,
          text,
          textType,
          textColor,
          textItalic,
          textDecoration,
          textBreak,
          textTruncate,
          textWeight
        )} 
        style={style}
      >
        {children}
      </div>
    );
  }
  export default CardHeader;