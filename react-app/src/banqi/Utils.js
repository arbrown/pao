
var PieceStrength = ["Q","P","H","C","E","G","K"]

function StrengthCompareD(a,b){
  return StrengthCompare(a,b,true)
};

function StrengthCompare(a, b, desc){
  a = a.toUpperCase();
  b = b.toUpperCase();
  let aStrength = PieceStrength.indexOf(a);
  let bStrength = PieceStrength.indexOf(b);
  if (!(aStrength > -1 && bStrength > -1)){
    return 0;
  }
  if (desc){
    return bStrength - aStrength;
  }
  return aStrength - bStrength;
};

const NotationToColor = {
    'K' : 'black',
    'G' : 'black',
    'E' : 'black',
    'C' : 'black',
    'H' : 'black',
    'P' : 'black',
    'Q' : 'black',
    'k' : 'red',
    'g' : 'red',
    'e' : 'red',
    'c' : 'red',
    'h' : 'red',
    'p' : 'red',
    'q' : 'red',
  };
  
const NotationToCss = {
    'K' : 'black-king',
    'G' : 'black-guard',
    'E' : 'black-elephant',
    'C' : 'black-cart',
    'H' : 'black-horse',
    'P' : 'black-pawn',
    'Q' : 'black-cannon',
    'k' : 'red-king',
    'g' : 'red-guard',
    'e' : 'red-elephant',
    'c' : 'red-cart',
    'h' : 'red-horse',
    'p' : 'red-pawn',
    'q' : 'red-cannon',
    '?' : 'unflipped-piece'
  };
  

export { StrengthCompare, StrengthCompareD, NotationToColor, NotationToCss }