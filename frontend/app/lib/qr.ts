const VERSION = 3
const SIZE = 29
const DATA_CODEWORDS = 55
const ECC_CODEWORDS = 15
const MASK_PATTERN = 0
const ERROR_CORRECTION_LOW = 1

type Matrix = Array<Array<boolean | null>>

const gfExp = new Array<number>(512)
const gfLog = new Array<number>(256).fill(0)

let gfValue = 1
for (let i = 0; i < 255; i++) {
  gfExp[i] = gfValue
  gfLog[gfValue] = i
  gfValue <<= 1
  if (gfValue & 0x100) {
    gfValue ^= 0x11d
  }
}
for (let i = 255; i < 512; i++) {
  gfExp[i] = gfExp[i - 255]!
}

export const createQrDataUrl = (value: string, size = 240) => {
  const svg = createQrSvg(value, size)
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`
}

export const createQrSvg = (value: string, size = 240) => {
  const codewords = createCodewords(value)
  const matrix = createEmptyMatrix()
  const reserved = createReservedMatrix()

  drawFunctionPatterns(matrix, reserved)
  drawCodewords(matrix, reserved, codewords)
  drawFormatBits(matrix, reserved)

  const quiet = 4
  const viewSize = SIZE + quiet * 2
  const rects: string[] = []

  for (let y = 0; y < SIZE; y++) {
    for (let x = 0; x < SIZE; x++) {
      if (matrix[y]?.[x]) {
        rects.push(`<rect x="${x + quiet}" y="${y + quiet}" width="1" height="1"/>`)
      }
    }
  }

  return `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 ${viewSize} ${viewSize}" shape-rendering="crispEdges"><rect width="${viewSize}" height="${viewSize}" fill="#f8fafc"/><g fill="#111">${rects.join('')}</g></svg>`
}

const createCodewords = (value: string) => {
  const bytes = Array.from(new TextEncoder().encode(value))
  const capacityBits = DATA_CODEWORDS * 8
  const bits: number[] = []

  if (bytes.length > 53) {
    throw new Error('QR payload is too long for Mautrade deposit QR')
  }

  appendBits(bits, 0b0100, 4)
  appendBits(bits, bytes.length, 8)
  for (const byte of bytes) {
    appendBits(bits, byte, 8)
  }

  const terminator = Math.min(4, capacityBits - bits.length)
  appendBits(bits, 0, terminator)
  while (bits.length % 8 !== 0) {
    bits.push(0)
  }

  const data: number[] = []
  for (let i = 0; i < bits.length; i += 8) {
    data.push(bitsToByte(bits.slice(i, i + 8)))
  }

  for (let pad = 0; data.length < DATA_CODEWORDS; pad++) {
    data.push(pad % 2 === 0 ? 0xec : 0x11)
  }

  return [...data, ...reedSolomonRemainder(data, ECC_CODEWORDS)]
}

const appendBits = (bits: number[], value: number, length: number) => {
  for (let i = length - 1; i >= 0; i--) {
    bits.push((value >>> i) & 1)
  }
}

const bitsToByte = (bits: number[]) => {
  return bits.reduce((value, bit) => (value << 1) | bit, 0)
}

const reedSolomonRemainder = (data: number[], degree: number) => {
  const generator = reedSolomonGenerator(degree)
  const result = [...data, ...new Array<number>(degree).fill(0)]

  for (let i = 0; i < data.length; i++) {
    const factor = result[i] ?? 0
    if (factor === 0) continue
    for (let j = 0; j < generator.length; j++) {
      result[i + j] = (result[i + j] ?? 0) ^ gfMultiply(generator[j]!, factor)
    }
  }

  return result.slice(data.length)
}

const reedSolomonGenerator = (degree: number) => {
  let result = [1]
  for (let i = 0; i < degree; i++) {
    result = polynomialMultiply(result, [1, gfExp[i]!])
  }
  return result
}

const polynomialMultiply = (left: number[], right: number[]) => {
  const result = new Array<number>(left.length + right.length - 1).fill(0)
  for (let i = 0; i < left.length; i++) {
    for (let j = 0; j < right.length; j++) {
      result[i + j] = (result[i + j] ?? 0) ^ gfMultiply(left[i]!, right[j]!)
    }
  }
  return result
}

const gfMultiply = (left: number, right: number) => {
  if (left === 0 || right === 0) return 0
  return gfExp[gfLog[left]! + gfLog[right]!]!
}

const createEmptyMatrix = (): Matrix => {
  return Array.from({ length: SIZE }, () => new Array<boolean | null>(SIZE).fill(null))
}

const createReservedMatrix = () => {
  return Array.from({ length: SIZE }, () => new Array<boolean>(SIZE).fill(false))
}

const drawFunctionPatterns = (matrix: Matrix, reserved: boolean[][]) => {
  drawFinder(matrix, reserved, 0, 0)
  drawFinder(matrix, reserved, SIZE - 7, 0)
  drawFinder(matrix, reserved, 0, SIZE - 7)
  drawAlignment(matrix, reserved, 22, 22)

  for (let i = 8; i < SIZE - 8; i++) {
    setFunction(matrix, reserved, i, 6, i % 2 === 0)
    setFunction(matrix, reserved, 6, i, i % 2 === 0)
  }

  setFunction(matrix, reserved, 8, 4 * VERSION + 9, true)
  reserveFormatAreas(matrix, reserved)
}

const drawFinder = (matrix: Matrix, reserved: boolean[][], x: number, y: number) => {
  for (let dy = -1; dy <= 7; dy++) {
    for (let dx = -1; dx <= 7; dx++) {
      const xx = x + dx
      const yy = y + dy
      if (!inBounds(xx, yy)) continue

      const inPattern = dx >= 0 && dx <= 6 && dy >= 0 && dy <= 6
      const dark = inPattern && (dx === 0 || dx === 6 || dy === 0 || dy === 6 || (dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4))
      setFunction(matrix, reserved, xx, yy, dark)
    }
  }
}

const drawAlignment = (matrix: Matrix, reserved: boolean[][], centerX: number, centerY: number) => {
  for (let dy = -2; dy <= 2; dy++) {
    for (let dx = -2; dx <= 2; dx++) {
      const distance = Math.max(Math.abs(dx), Math.abs(dy))
      setFunction(matrix, reserved, centerX + dx, centerY + dy, distance === 2 || distance === 0)
    }
  }
}

const reserveFormatAreas = (matrix: Matrix, reserved: boolean[][]) => {
  for (let i = 0; i <= 8; i++) {
    reserve(matrix, reserved, 8, i)
    reserve(matrix, reserved, i, 8)
  }
  for (let i = 0; i < 8; i++) {
    reserve(matrix, reserved, SIZE - 1 - i, 8)
  }
  for (let i = 8; i < 15; i++) {
    reserve(matrix, reserved, 8, SIZE - 15 + i)
  }
}

const drawCodewords = (matrix: Matrix, reserved: boolean[][], codewords: number[]) => {
  const bits = codewords.flatMap(codeword => Array.from({ length: 8 }, (_, index) => (codeword >>> (7 - index)) & 1))
  let bitIndex = 0
  let upward = true

  for (let right = SIZE - 1; right >= 1; right -= 2) {
    if (right === 6) right--

    for (let vertical = 0; vertical < SIZE; vertical++) {
      const y = upward ? SIZE - 1 - vertical : vertical
      for (let column = 0; column < 2; column++) {
        const x = right - column
        if (reserved[y]?.[x]) continue

        let dark = bitIndex < bits.length ? bits[bitIndex] === 1 : false
        if ((x + y) % 2 === 0) {
          dark = !dark
        }
        matrix[y]![x] = dark
        bitIndex++
      }
    }

    upward = !upward
  }
}

const drawFormatBits = (matrix: Matrix, reserved: boolean[][]) => {
  const bits = formatBits(ERROR_CORRECTION_LOW, MASK_PATTERN)
  const bit = (index: number) => ((bits >>> index) & 1) === 1

  for (let i = 0; i <= 5; i++) {
    setFunction(matrix, reserved, 8, i, bit(i))
  }
  setFunction(matrix, reserved, 8, 7, bit(6))
  setFunction(matrix, reserved, 8, 8, bit(7))
  setFunction(matrix, reserved, 7, 8, bit(8))
  for (let i = 9; i < 15; i++) {
    setFunction(matrix, reserved, 14 - i, 8, bit(i))
  }

  for (let i = 0; i < 8; i++) {
    setFunction(matrix, reserved, SIZE - 1 - i, 8, bit(i))
  }
  for (let i = 8; i < 15; i++) {
    setFunction(matrix, reserved, 8, SIZE - 15 + i, bit(i))
  }
}

const formatBits = (errorCorrectionLevel: number, mask: number) => {
  const data = (errorCorrectionLevel << 3) | mask
  let remainder = data << 10
  const generator = 0x537

  for (let i = 14; i >= 10; i--) {
    if (((remainder >>> i) & 1) !== 0) {
      remainder ^= generator << (i - 10)
    }
  }

  return ((data << 10) | remainder) ^ 0x5412
}

const reserve = (matrix: Matrix, reserved: boolean[][], x: number, y: number) => {
  if (!inBounds(x, y)) return
  reserved[y]![x] = true
  if (matrix[y]![x] === null) {
    matrix[y]![x] = false
  }
}

const setFunction = (matrix: Matrix, reserved: boolean[][], x: number, y: number, dark: boolean) => {
  if (!inBounds(x, y)) return
  matrix[y]![x] = dark
  reserved[y]![x] = true
}

const inBounds = (x: number, y: number) => x >= 0 && y >= 0 && x < SIZE && y < SIZE
