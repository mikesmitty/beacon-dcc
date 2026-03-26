package dcc

/*
// saves DC frequency (0..3) in spare functions 29,30,31
void DCC::setDCFreq(int loco,byte freq) {
  if (loco==0 || freq>3) return;
  auto reg=lookupSpeedTable(loco,true);
  // drop and replace F29,30,31 (top 3 bits)
  auto newFunctions=speedTable[reg].functions & 0x1FFFFFFFUL;
  if (freq==1)      newFunctions |= (1UL<<29); // F29
  else if (freq==2) newFunctions |= (1UL<<30); // F30
  else if (freq==3) newFunctions |= (1UL<<31); // F31
  if (newFunctions==speedTable[reg].functions) return; // no change
  speedTable[reg].functions=newFunctions;
  CommandDistributor::broadcastLoco(reg);
}
*/
