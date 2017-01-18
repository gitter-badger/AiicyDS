
# where the application code will be generated
# relative to the "tools" directory
APP_ROOT = "../app"



ARCHITECTURE = {
    'controllers' : {
        'Module1' : {},
        'Articles' : {
            'description': 'a module that does something',
            'methods' : {
                'show' : {},
                'do_something_else' : {}
            },
            'forms' : {
                'add_comment' : {}
            }
        }
    },

    'models': {
        'MyModel' : {}
    },

    'models_controllers': [
        ('MyModel','Module1'),
        ('MyModel','Articles')

    ]



}



REPLACEMENTS = {
    '@AUTHOR@' : 'Allan',
    '@EMAIL@': 'Your_email.com',
    '@PROJECT_NAME_CODE@' : 'MySuiWiki',
    '@PROJECT_NAME_HUMAN@': 'MySui wiki',
    '@PROJECT_NS@': 'mysuiwiki',
    '@MAIN_CLASS@' : 'MySuiWiki',
    '@MAIN_CLASS_HEADER@' : 'MYSUIWIKI',
    '@DESCRIPTION@' : 'Description of your project',
    '@PROJECT_WEBSITE' : 'link to your project',
    '@YEARS@' : 'copyright years'
}


