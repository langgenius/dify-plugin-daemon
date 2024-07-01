def main():
    while True:
        try:
            x = int(input('Enter a number: '))
            y = int(input('Enter another number: '))
            print(f'{x} + {y} = {x + y}')
        except ValueError:
            print('Invalid input. Please enter a number.')

if __name__ == '__main__':
    main()