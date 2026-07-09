from machine import Pin

class _Input:
    def __init__(self):
        self.pin_a = Pin(4, Pin.IN, Pin.PULL_UP) #enable internal pull-up resistor
        self.pin_b = Pin(5, Pin.IN, Pin.PULL_UP) 
        self.dpad_up = Pin(9, Pin.IN, Pin.PULL_UP)
        self.dpad_left = Pin(10, Pin.IN, Pin.PULL_UP)
        self.dpad_right = Pin(11, Pin.IN, Pin.PULL_UP)
        self.dpad_down = Pin(12, Pin.IN, Pin.PULL_UP)
        self.keysHeld = {
            "A": False,
            "B": False,
            "UP": False,
            "DOWN": False,
            "LEFT": False,
            "RIGHT": False
        }
        self.oldKeysHeld = self.keysHeld.copy()  
    def update(self):
        self.oldKeysHeld = self.keysHeld.copy() 
        self.keysHeld["A"] = not self.pin_a.value() #invert the value because of pull-up resistor
        self.keysHeld["B"] = not self.pin_b.value()
        self.keysHeld["UP"] = not self.dpad_up.value()
        self.keysHeld["DOWN"] = not self.dpad_down.value()
        self.keysHeld["LEFT"] = not self.dpad_left.value()
        self.keysHeld["RIGHT"] = not self.dpad_right.value()

    def is_pressed(self, key):
        return self.keysHeld.get(key, False)
    def was_just_pressed(self, key):
        return self.keysHeld.get(key, False) and not self.oldKeysHeld.get(key, False)
    def was_just_released(self, key):
        return not self.keysHeld.get(key, False) and self.oldKeysHeld.get(key, False)

    
_instance = _Input()

update = _instance.update
is_pressed = _instance.is_pressed
was_just_pressed = _instance.was_just_pressed
was_just_released = _instance.was_just_released