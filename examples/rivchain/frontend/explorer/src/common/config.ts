const {
  VUE_APP_API_URL,
  VUE_APP_PRECISION,
  VUE_APP_UNIT,
  VUE_APP_NAME
} = process.env

export const API_URL =
  VUE_APP_API_URL || 'https://explorer.testnet.nbh-digital.com'
export const NAME = VUE_APP_NAME || 'Goldchain'
export const PRECISION = parseInt(VUE_APP_PRECISION) || 9
export const UNIT = VUE_APP_UNIT || 'GFT'
