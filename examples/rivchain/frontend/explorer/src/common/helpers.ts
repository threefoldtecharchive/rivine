import { PRECISION } from './config'
import {
  BlockstakeOutputInfo, CoinOutputInfo, ConditionType,
  UnlockhashCondition, AtomicSwapCondition, TimelockCondition, Condition,
  MultisignatureCondition
} from 'rivine-ts-types'

export function toLocalDecimalNotation (x: any) {
  if (!x) return
  const w = parseFloat(x)
  const v = Number(w)
  return v.toLocaleString(navigator.language, {
    maximumFractionDigits: PRECISION
  })
}

export function formatReadableDate (time: number) {
  const blockDate = new Date(time * 1000)
  const day = blockDate.getDate()
  const month = blockDate.toLocaleString('default', { month: 'long' })
  const year = blockDate.getFullYear()
  const hours = blockDate.getHours()
  const tempMinutes = blockDate.getMinutes()
  const minutes = tempMinutes < 10 ? `0${tempMinutes}` : tempMinutes

  return `${hours}:${minutes}, ${month} ${day}, ${year}`
}

// formatTimeElapsed takes a duration in seconds and returns it in a more human readable format
export function formatTimeElapsed (seconds: number) {
  if (!seconds) {
    return '0 seconds'
  }
  const levels = [
    [Math.floor(seconds / 31536000), 'years'],
    [Math.floor((seconds % 31536000) / 86400), 'days'],
    [Math.floor(((seconds % 31536000) % 86400) / 3600), 'hours'],
    [Math.floor((((seconds % 31536000) % 86400) % 3600) / 60), 'minutes'],
    [(((seconds % 31536000) % 86400) % 3600) % 60, 'seconds']
  ]
  let returntext = ''

  // tslint:disable-next-line
  for (let i = 0, max = levels.length; i < max; i++) {
    if (levels[i][0] === 0) continue
    // tslint:disable-next-line
    returntext += ', ' + levels[i][0] + ' ' + (levels[i][0] === 1 ? levels[i][1].toString().substr(0, levels[i][1].toString().length-1): levels[i][1])
  }
  return returntext.trim().substr(2)
}

export function formatReadableDateForCharts (time: number) {
  const blockDate = new Date(time * 1000)
  return blockDate
}

export function getUnlockhashFromCondition (condition: Condition): string {
  switch (condition.getConditionType()) {
    case ConditionType.NilCondition:
      return ''
    case ConditionType.UnlockhashCondition:
      const uhCondition = condition as UnlockhashCondition
      return uhCondition.unlockhash
    case ConditionType.AtomicSwapCondition:
      const atCondition = condition as AtomicSwapCondition
      return atCondition.receiver
    case ConditionType.TimelockCondition:
      const tmCondition = condition as TimelockCondition
      return getUnlockhashFromCondition(tmCondition)
    case ConditionType.MultisignatureCondition:
      const msCondition = condition as MultisignatureCondition
      return msCondition.unlockhashes.join(',')
    default:
      return ''
  }
}

export function getUnlockHash (outputInfo: BlockstakeOutputInfo | CoinOutputInfo): string {
  debugger
  if (outputInfo.output.isBlockCreatorReward) {
    if (outputInfo.output.unlockhash) {
      return outputInfo.output.unlockhash
    }
  } else if (outputInfo.output.isCustodyFee) {
    if (outputInfo.output.condition.custodyFeeVoidAddress) {
      return outputInfo.output.condition.custodyFeeVoidAddress
    }
  } else {
    if (outputInfo.output.condition) {
      return getUnlockhashFromCondition(outputInfo.output.condition)
    }
  }
  return ''
}
